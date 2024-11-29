package realy

import (
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
	Cancel               context.F
	options              *Options
	relay                relay.I
	clientsMu            sync.Mutex
	clients              map[*websocket.Conn]struct{}
	Addr                 S
	serveMux             *http.ServeMux
	httpServer           *http.Server
	authRequired         bool
	maxLimit             N
	adminUser, adminPass S
}

func (s *Server) Router() *http.ServeMux { return s.serveMux }

type ServerParams struct {
	Ctx
	Cancel               context.F
	Rl                   relay.I
	DbPath               S
	MaxLimit             N
	AdminUser, AdminPass S
}

// NewServer initializes the realy and its storage using their respective Init methods,
// returning any non-nil errors, and returns a Server ready to listen for HTTP requests.
func NewServer(sp ServerParams, opts ...Option) (*Server, E) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	var authRequired bool
	if ar, ok := sp.Rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	srv := &Server{
		Ctx:          sp.Ctx,
		Cancel:       sp.Cancel,
		relay:        sp.Rl,
		clients:      make(map[*websocket.Conn]struct{}),
		serveMux:     http.NewServeMux(),
		options:      options,
		authRequired: authRequired,
		maxLimit:     sp.MaxLimit,
		adminUser:    sp.AdminUser,
		adminPass:    sp.AdminPass,
	}

	if storage := sp.Rl.Storage(context.Bg()); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}

	// init the relay
	if err := sp.Rl.Init(); chk.T(err) {
		return nil, fmt.Errorf("realy init: %w", err)
	}

	// start listening from events from other sources, if any
	if inj, ok := sp.Rl.(relay.Injector); ok {
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
		return
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		s.HandleNIP11(w, r)
		return
	}
	s.HandleAdmin(w, r)
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
		WriteTimeout: 7 * time.Second,
		ReadTimeout:  7 * time.Second,
		IdleTimeout:  28 * time.Second,
	}
	// notify caller that we're starting
	for _, startedC := range started {
		close(startedC)
	}
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
	s.relay.Storage(s.Ctx).Close()
	log.W.Ln("shutting down relay listener")
	s.httpServer.Shutdown(s.Ctx)
	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(s.Ctx)
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
