package realy

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"

	realy_lol "realy.lol"
	"realy.lol/context"
	"realy.lol/realy/listeners"
	"realy.lol/realy/options"
	"realy.lol/relay"
	"realy.lol/signer"
	"realy.lol/store"
)

type Server struct {
	Ctx            context.T
	Cancel         context.F
	options        *options.T
	relay          relay.I
	clientsMu      sync.Mutex
	clients        map[*websocket.Conn]struct{}
	Addr           string
	mux            *ServeMux
	httpServer     *http.Server
	authRequired   bool
	publicReadable bool
	maxLimit       int
	admins         []signer.I
	owners         [][]byte
	Listeners      *listeners.T
	huma.API
	ConfigurationMx sync.Mutex
	Configuration   *store.Configuration
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
		authRequired = ar.AuthEnabled()
	}
	if storage := sp.Rl.Storage(); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}
	serveMux := NewServeMux()
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
		Listeners:      listeners.New(sp.Ctx),
		API: NewHuma(serveMux, sp.Rl.Name(), realy_lol.Version,
			realy_lol.Description),
	}
	huma.AutoRegister(s.API, NewEvent(s))
	huma.AutoRegister(s.API, NewFilter(s))
	huma.AutoRegister(s.API, NewEvents(s))
	huma.AutoRegister(s.API, NewSubscribe(s))

	huma.AutoRegister(s.API, NewExport(s))
	huma.AutoRegister(s.API, NewImport(s))

	huma.AutoRegister(s.API, NewRescan(s))
	huma.AutoRegister(s.API, NewShutdown(s))
	huma.AutoRegister(s.API, NewDisconnect(s))
	huma.AutoRegister(s.API, NewConfiguration(s))
	huma.AutoRegister(s.API, NewNuke(s))

	// load configuration if it has been set
	if c, ok := s.relay.Storage().(store.Configurationer); ok {
		s.ConfigurationMx.Lock()
		if s.Configuration, err = c.GetConfiguration(); chk.E(err) {
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
				s.Listeners.NotifyListeners(s.authRequired, s.publicReadable, ev)
			}
		}()
	}
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote := GetRemoteFromReq(r)
	for _, a := range s.Configuration.BlockList {
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
	log.I.F("http request: %s from %s", r.URL.String(), GetRemoteFromReq(r))
	s.mux.ServeHTTP(w, r)
	// s.HandleHTTP(w, r)
	// s.mux.ServeHTTP(w, r)
}

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

func (s *Server) Router() *http.ServeMux {
	return s.mux.ServeMux
}

func fprintf(w io.Writer, format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
