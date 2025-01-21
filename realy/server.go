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
	cx
	sync.Mutex
	relay.I
	*http.ServeMux
	*listeners.T
	*options.O
	cancel       context.F
	clients      map[*websocket.Conn]struct{}
	Addr         st
	httpServer   *http.Server
	authRequired bo
	maxLimit     no
	adminUser    st
	adminPass    st
}

type ServerParams struct {
	Ctx    cx
	Cancel context.F
	relay.I
	DbPath               st
	MaxLimit             no
	AdminUser, AdminPass st
}

func NewServer(sp ServerParams, opts ...options.F) (*Server, er) {
	op := options.Default()
	for _, opt := range opts {
		opt(op)
	}
	var authRequired bo
	if ar, ok := sp.I.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	srv := &Server{
		cx:           sp.Ctx,
		cancel:       sp.Cancel,
		I:            sp.I,
		clients:      make(map[*websocket.Conn]struct{}),
		ServeMux:     http.NewServeMux(),
		O:            op,
		authRequired: authRequired,
		maxLimit:     sp.MaxLimit,
		adminUser:    sp.AdminUser,
		adminPass:    sp.AdminPass,
		T:            listeners.New(),
	}
	if storage := sp.Storage(); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}
	if err := sp.Init(); chk.T(err) {
		return nil, fmt.Errorf("realy init: %w", err)
	}
	if inj, ok := sp.I.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				srv.NotifyListeners(srv.authRequired, ev)
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
		// WriteTimeout: 7 * time.Second,
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
	s.cancel()
	s.Lock()
	defer s.Unlock()
	for c := range s.clients {
		log.I.Ln("disconnecting", c.RemoteAddr())
		chk.E(c.WriteControl(websocket.CloseMessage, nil,
			time.Now().Add(time.Second)))
		chk.E(c.Close())
		delete(s.clients, c)
	}
	log.W.Ln("closing event store")
	chk.E(s.Storage().Close())
	log.W.Ln("shutting down relay listener")
	chk.E(s.httpServer.Shutdown(s.cx))
	if f, ok := s.I.(relay.ShutdownAware); ok {
		f.OnShutdown(s.cx)
	}
}

func (s *Server) Router() *http.ServeMux { return s.ServeMux }

func fprintf(w io.Writer, format st, a ...any) {
	_, _ = fmt.Fprintf(w, format,
		a...)
}
