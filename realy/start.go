package realy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
	"realy.lol/event"
	"realy.lol/relay"
)

// Server is a base for package users to implement nostr relays.
// It can serve HTTP requests and websockets, passing control over to a relay implementation.
//
// To implement a relay, it is enough to satisfy [Relay] interface. Other interfaces are
// [Informationer], [CustomWebSocketHandler], [ShutdownAware] and AdvancedXxx types.
// See their respective doc comments.
//
// The basic usage is to call Start or StartConf, which starts serving immediately.
// For a more fine-grained control, use NewServer.
// See [basic/main.go], [whitelisted/main.go], [expensive/main.go] and [rss-bridge/main.go]
// for example implementations.
//
// The following resource is a good starting point for details on what nostr protocol is
// and how it works: https://github.com/nostr-protocol/nostr
type Server struct {
	options *Options
	relay   relay.I
	// keep a connection reference to all connected clients for Server.Shutdown
	clientsMu    sync.Mutex
	clients      map[*websocket.Conn]struct{}
	Addr         S
	serveMux     *http.ServeMux
	httpServer   *http.Server
	authRequired bool
}

func (s *Server) Router() *http.ServeMux {
	return s.serveMux
}

// NewServer initializes the realy and its storage using their respective Init methods,
// returning any non-nil errors, and returns a Server ready to listen for HTTP requests.
func NewServer(rl relay.I, dbPath S, opts ...Option) (*Server, E) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	srv := &Server{
		relay:    rl,
		clients:  make(map[*websocket.Conn]struct{}),
		serveMux: &http.ServeMux{},
		options:  options,
	}

	if storage := rl.Storage(context.Background()); storage != nil {
		if err := storage.Init(dbPath); err != nil {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}

	// init the realy
	if err := rl.Init(); err != nil {
		return nil, fmt.Errorf("realy init: %w", err)
	}

	// start listening from events from other sources, if any
	if inj, ok := rl.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				notifyListeners(ev)
			}
		}()
	}

	return srv, nil
}

// ServeHTTP implements http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		s.HandleNIP11(w, r)
	} else {
		s.serveMux.ServeHTTP(w, r)
	}
}

func (s *Server) Start(host S, port int, started ...chan bool) E {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	log.I.F("starting relay listener at %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.Addr = ln.Addr().String()
	s.httpServer = &http.Server{
		Handler:      cors.Default().Handler(s),
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// notify caller that we're starting
	for _, started := range started {
		close(started)
	}

	if err = s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

// Shutdown sends a websocket close control message to all connected clients.
//
// If the realy is ShutdownAware, Shutdown calls its OnShutdown, passing the context as is.
// Note that the HTTP server make some time to shutdown and so the context deadline,
// if any, may have been shortened by the time OnShutdown is called.
func (s *Server) Shutdown(ctx context.Context) {
	s.httpServer.Shutdown(ctx)

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(time.Second))
		conn.Close()
		delete(s.clients, conn)
	}

	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(ctx)
	}
}

type Option func(*Options)

type Options struct {
	perConnectionLimiter *rate.Limiter
	skipEventFunc        func(*event.T) bool
}

func DefaultOptions() *Options {
	return &Options{}
}

func WithPerConnectionLimiter(rps rate.Limit, burst N) Option {
	return func(o *Options) {
		o.perConnectionLimiter = rate.NewLimiter(rps, burst)
	}
}

func WithSkipEventFunc(skipEventFunc func(*event.T) bool) Option {
	return func(o *Options) {
		o.skipEventFunc = skipEventFunc
	}
}
