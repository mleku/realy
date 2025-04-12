package realy

import (
	"net/http"
	"time"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/realy/options"
	"realy.mleku.dev/realy/subscribers"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/store"
)

func (s *Server) AdminAuth(r *http.Request,
	tolerance ...time.Duration) (authed bool,
	pubkey []byte) {

	return s.adminAuth(r, tolerance...)
}

func (s *Server) Storage() store.I { return s.relay.Storage() }

func (s *Server) Configuration() store.Configuration {
	s.ConfigurationMx.Lock()
	defer s.ConfigurationMx.Unlock()
	return *s.configuration
}

func (s *Server) SetConfiguration(cfg *store.Configuration) {
	s.ConfigurationMx.Lock()
	s.configuration = cfg
	s.ConfigurationMx.Unlock()
}

func (s *Server) Relay() relay.I { return s.relay }

func (s *Server) Disconnect() { s.disconnect() }

func (s *Server) AddEvent(
	c context.T, rl relay.I, ev *event.T, hr *http.Request, origin string,
	authedPubkey []byte) (accepted bool, message []byte) {

	return s.addEvent(c, rl, ev, hr, origin, authedPubkey)
}

func (s *Server) AcceptEvent(
	c context.T, ev *event.T, hr *http.Request, origin string,
	authedPubkey []byte) (accept bool, notice string, afterSave func()) {

	return s.relay.AcceptEvent(c, ev, hr, origin, authedPubkey)
}

func (s *Server) Listeners() *subscribers.S { return s.listeners }

func (s *Server) PublicReadable() bool { return s.publicReadable }

func (s *Server) Context() context.T { return s.Ctx }

func (s *Server) Owners() [][]byte { return s.owners }

func (s *Server) AuthRequired() bool { return s.authRequired }

func (s *Server) Options() *options.T { return s.options }

var _ interfaces.Server = &Server{}
