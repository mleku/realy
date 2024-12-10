package realy

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/rs/cors"

	"realy.lol/context"
	"realy.lol/realy/listeners"
	"realy.lol/realy/options"
	"realy.lol/relay"
)

type Server struct {
	Ctx                  cx
	Cancel               context.F
	options              *options.T
	relay                relay.I
	clientsMu            sync.Mutex
	clients              map[*websocket.Conn]struct{}
	Addr                 st
	serveMux             *http.ServeMux
	httpServer           *http.Server
	authRequired         bo
	maxLimit             no
	adminUser, adminPass st
	listeners            *listeners.T
}

type ServerParams struct {
	Ctx                  cx
	Cancel               context.F
	Rl                   relay.I
	DbPath               st
	MaxLimit             no
	AdminUser, AdminPass st
}

func NewServer(sp ServerParams, opts ...options.O) (*Server, er) {
	op := options.Default()
	for _, opt := range opts {
		opt(op)
	}
	var authRequired bo
	if ar, ok := sp.Rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	srv := &Server{
		Ctx:          sp.Ctx,
		Cancel:       sp.Cancel,
		relay:        sp.Rl,
		clients:      make(map[*websocket.Conn]struct{}),
		serveMux:     http.NewServeMux(),
		options:      op,
		authRequired: authRequired,
		maxLimit:     sp.MaxLimit,
		adminUser:    sp.AdminUser,
		adminPass:    sp.AdminPass,
		listeners:    listeners.New(),
	}
	if storage := sp.Rl.Storage(context.Bg()); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}
	if err := sp.Rl.Init(); chk.T(err) {
		return nil, fmt.Errorf("realy init: %w", err)
	}
	if inj, ok := sp.Rl.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				srv.listeners.NotifyListeners(srv.authRequired, ev)
			}
		}()
	}
	return srv, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		s.handleWebsocket(w, r)
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		s.handleRelayInfo(w, r)
	} else {
		s.handleAdmin(w, r)
	}
}

func (s *Server) Start(host st, port int, started ...chan bo) er {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	log.I.F("starting relay listener at %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Addr = ln.Addr().String()
	s.httpServer = &http.Server{Handler: cors.Default().Handler(s), Addr: addr,
		//WriteTimeout: 7 * time.Second,
		ReadHeaderTimeout: 7 * time.Second,
		IdleTimeout:       28 * time.Second}
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
	chk.E(s.relay.Storage(s.Ctx).Close())
	log.W.Ln("shutting down relay listener")
	chk.E(s.httpServer.Shutdown(s.Ctx))
	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(s.Ctx)
	}
}

func (s *Server) Router() *http.ServeMux {
	return s.serveMux
}

func fprintf(w io.Writer, format st, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
