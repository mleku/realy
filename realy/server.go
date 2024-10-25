package realy

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
	"realy.lol/context"
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
type Server struct {
	Ctx
	Cancel                  context.F
	options                 *Options
	relay                   relay.I
	clientsMu               sync.Mutex
	clients                 map[*websocket.Conn]struct{}
	Addr, AdminAddr         S
	serveMux                *http.ServeMux
	httpServer, adminServer *http.Server
	authRequired            bool
}

func (s *Server) Router() *http.ServeMux { return s.serveMux }

// NewServer initializes the realy and its storage using their respective Init methods,
// returning any non-nil errors, and returns a Server ready to listen for HTTP requests.
func NewServer(c Ctx, cancel context.F, rl relay.I, dbPath S, opts ...Option) (*Server, E) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	var authRequired bool
	if ar, ok := rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	srv := &Server{
		Ctx:          c,
		Cancel:       cancel,
		relay:        rl,
		clients:      make(map[*websocket.Conn]struct{}),
		serveMux:     http.NewServeMux(),
		options:      options,
		authRequired: authRequired,
	}

	if storage := rl.Storage(context.Bg()); storage != nil {
		if err := storage.Init(dbPath); err != nil {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}

	// init the relay
	if err := rl.Init(); err != nil {
		return nil, fmt.Errorf("realy init: %w", err)
	}

	// start listening from events from other sources, if any
	if inj, ok := rl.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				notifyListeners(srv.authRequired, ev)
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
	} else if s.AdminAddr == r.Host ||
		strings.HasPrefix(s.AdminAddr, "127.0.0.1") &&
			strings.HasPrefix(r.Host, "localhost") {
		s.HandleAdmin(w, r)
	} else {
		s.serveMux.ServeHTTP(w, r)
	}
}

func (s *Server) Start(host S, port int, adminHost S, adminPort int, started ...chan bool) E {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	log.I.F("starting relay listener at %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	adminAddr := net.JoinHostPort(adminHost, strconv.Itoa(adminPort))
	log.I.F("starting relay admin listener at %s", adminAddr)
	var aln net.Listener
	aln, err = net.Listen("tcp", adminAddr)
	if err != nil {
		return err
	}
	s.Addr = ln.Addr().String()
	s.AdminAddr = aln.Addr().String()
	s.httpServer = &http.Server{
		Handler:      cors.Default().Handler(s),
		Addr:         addr,
		WriteTimeout: 7 * time.Second,
		ReadTimeout:  7 * time.Second,
		IdleTimeout:  28 * time.Second,
	}
	s.adminServer = &http.Server{
		Handler: cors.Default().Handler(s),
		Addr:    adminAddr,
		// WriteTimeout: 4 * time.Second,
		// ReadTimeout:  4 * time.Second,
		// IdleTimeout:  30 * time.Second,
	}

	// notify caller that we're starting
	for _, startedC := range started {
		close(startedC)
	}

	go func() {
		if err = s.adminServer.Serve(aln); errors.Is(err, http.ErrServerClosed) {
		}
	}()
	if err = s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
	} else if err != nil {
	}
	return nil
}

// Shutdown sends a websocket close control message to all connected clients.
//
// If the realy is ShutdownAware, Shutdown calls its OnShutdown, passing the context as is.
// Note that the HTTP server make some time to shutdown and so the context deadline,
// if any, may have been shortened by the time OnShutdown is called.
func (s *Server) Shutdown() {
	c := s.Ctx
	log.I.Ln("shutting down relay")
	s.Cancel()
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		log.I.Ln("disconnecting", conn.RemoteAddr())
		conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(time.Second))
		conn.Close()
		delete(s.clients, conn)
	}
	log.W.Ln("closing event store")
	s.relay.Storage(c).Close()
	log.W.Ln("shutting down relay listener")
	s.httpServer.Shutdown(c)
	log.W.S("shutting down admin listener")
	s.adminServer.Shutdown(c)
	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(c)
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
