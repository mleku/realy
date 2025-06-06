package realy

import (
	"net/http"
	"time"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/errorf"
	"realy.lol/event"
	"realy.lol/log"
	"realy.lol/realy/config"
	"realy.lol/realy/interfaces"
	"realy.lol/signer"
	"realy.lol/store"
)

func (s *Server) AdminAuth(r *http.Request, remote string,
	tolerance ...time.Duration) (authed bool, pubkey []byte) {

	return s.adminAuth(r, tolerance...)
}

func (s *Server) Storage() store.I { return s.Store }

func (s *Server) Configuration() config.C {
	s.configurationMx.Lock()
	defer s.configurationMx.Unlock()
	return s.configuration
}

func (s *Server) SetConfiguration(cfg config.C) (err error) {
	if len(cfg.Admins) == 0 {
		err = errorf.E("cannot set configuration without at least one admin")
		return
	}
	if c, ok := s.Store.(store.Configurationer); ok {
		chk.E(c.SetConfiguration(cfg))
		chk.E(s.UpdateConfiguration())
	}
	return err
}

func (s *Server) AddEvent(
	c context.T, ev *event.T, hr *http.Request, authedPubkey []byte,
	remote string) (accepted bool, message []byte) {

	return s.addEvent(c, ev, authedPubkey, remote)
}

func (s *Server) AcceptEvent(
	c context.T, ev *event.T, hr *http.Request, authedPubkey []byte,
	remote string) (accept bool, notice string, afterSave func()) {
	return s.acceptEvent(c, ev, authedPubkey, remote)
}

func (s *Server) PublicReadable() bool {
	s.configurationMx.Lock()
	defer s.configurationMx.Unlock()
	pr := s.configuration.PublicReadable
	log.T.F("public readable %v", pr)
	return pr
}

func (s *Server) Context() context.T { return s.Ctx }

func (s *Server) Owners() [][]byte {
	return s.owners
}

func (s *Server) SetOwners(owners [][]byte) {
	s.owners = owners
}

func (s *Server) AuthRequired() bool {
	s.configurationMx.Lock()
	defer s.configurationMx.Unlock()
	return s.configuration.AuthRequired
}

func (s *Server) OwnersFollowed(pubkey string) (ok bool) {
	_, ok = s.ownersFollowed[pubkey]
	return
}

func (s *Server) SetAdmins(admins []signer.I) {
	s.admins = admins
}

var _ interfaces.Server = &Server{}
