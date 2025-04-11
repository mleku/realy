package realy

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"

	realy_lol "realy.mleku.dev"
	"realy.mleku.dev/context"
	"realy.mleku.dev/openapi"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/realy/options"
	"realy.mleku.dev/realy/subscribers"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/signer"
	"realy.mleku.dev/store"
)

type Server struct {
	Ctx            context.T
	Cancel         context.F
	options        *options.T
	relay          relay.I
	clientsMu      sync.Mutex
	clients        map[*websocket.Conn]struct{}
	Addr           string
	mux            *openapi.ServeMux
	httpServer     *http.Server
	authRequired   bool
	publicReadable bool
	maxLimit       int
	admins         []signer.I
	owners         [][]byte
	listeners      *subscribers.S
	huma.API
	ConfigurationMx sync.Mutex
	configuration   *store.Configuration
}

type ServerParams struct {
	Ctx            context.T
	Cancel         context.F
	Rl             relay.I
	DbPath         string
	MaxLimit       int
	Admins         []signer.I
	Owners         [][]byte
	PublicReadable bool
}

func NewServer(sp *ServerParams, opts ...options.O) (s *Server, err error) {
	op := options.Default()
	for _, opt := range opts {
		opt(op)
	}
	var authRequired bool
	if ar, ok := sp.Rl.(relay.Authenticator); ok {
		authRequired = ar.AuthRequired()
	}
	if storage := sp.Rl.Storage(); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}
	serveMux := openapi.NewServeMux()
	s = &Server{
		Ctx:            sp.Ctx,
		Cancel:         sp.Cancel,
		relay:          sp.Rl,
		clients:        make(map[*websocket.Conn]struct{}),
		mux:            serveMux,
		options:        op,
		authRequired:   authRequired,
		publicReadable: sp.PublicReadable,
		maxLimit:       sp.MaxLimit,
		admins:         sp.Admins,
		owners:         sp.Rl.Owners(),
		listeners:      subscribers.New(sp.Ctx),
		API: openapi.NewHuma(serveMux, sp.Rl.Name(), realy_lol.Version,
			realy_lol.Description),
	}
	// register the http API operations
	huma.AutoRegister(s.API, openapi.NewOperations(s))
	// load configuration if it has been set
	if c, ok := s.relay.Storage().(store.Configurationer); ok {
		s.ConfigurationMx.Lock()
		if s.configuration, err = c.GetConfiguration(); chk.E(err) {
			s.configuration = &store.Configuration{}
		}
		s.ConfigurationMx.Unlock()
	}

	go func() {
		if err := s.relay.Init(); chk.E(err) {
			s.Shutdown()
		}
	}()
	if inj, ok := s.relay.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				s.listeners.NotifySubscribers(s.authRequired, s.publicReadable, ev)
			}
		}()
	}
	return s, nil
}

// ServeHTTP implements the relay's http handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote := helpers.GetRemoteFromReq(r)
	for _, a := range s.Configuration().BlockList {
		if strings.HasPrefix(remote, a) {
			log.W.F("rejecting request from %s because on blocklist", remote)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}
	// standard nostr protocol only governs the "root" path of the relay and websockets
	if r.URL.Path == "/" && r.Header.Get("Accept") == "application/nostr+json" {
		s.handleRelayInfo(w, r)
		return
	}
	if r.URL.Path == "/" && r.Header.Get("Upgrade") == "websocket" {
		s.handleWebsocket(w, r)
		return
	}
	log.I.F("http request: %s from %s", r.URL.String(), helpers.GetRemoteFromReq(r))
	s.mux.ServeHTTP(w, r)
	// s.HandleHTTP(w, r)
	// s.mux.ServeHTTP(w, r)
}

// Start up the relay.
func (s *Server) Start(host string, port int, started ...chan bool) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	log.I.F("starting relay listener at %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Addr = ln.Addr().String() // todo: this doesn't seem to do anything
	s.httpServer = &http.Server{
		Handler: cors.Default().Handler(s),
		Addr:    addr,
		// WriteTimeout: 7 * time.Second,
		ReadHeaderTimeout: 7 * time.Second,
		IdleTimeout:       28 * time.Second,
	}
	for _, startedC := range started {
		close(startedC)
	}
	if err = s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
	} else if err != nil {
	}
	return nil
}

// Shutdown the relay.
func (s *Server) Shutdown() {
	log.I.Ln("shutting down relay")
	s.Cancel()
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		log.I.Ln("disconnecting", conn.RemoteAddr())
		chk.E(conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(time.Second)))
		chk.E(conn.Close())
		delete(s.clients, conn)
	}
	log.W.Ln("closing event store")
	chk.E(s.relay.Storage().Close())
	log.W.Ln("shutting down relay listener")
	chk.E(s.httpServer.Shutdown(s.Ctx))
	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(s.Ctx)
	}
}

// Router returns the servemux that handles paths on the HTTP server of the relay.
func (s *Server) Router() *http.ServeMux {
	return s.mux.ServeMux
}
