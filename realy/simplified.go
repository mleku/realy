package realy

import (
	"realy.lol/realy/handler"
)

func init() {
	handler.RegisterCapability("/capabilities", "")
}

// CapabilitiesHandler handles a capabilities request.
func (s *Server) CapabilitiesHandler(h handler.H) {
	log.I.F("/capabilities request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/event", "")
}

// EventHandler handles a event storage request.
func (s *Server) EventHandler(h handler.H) {
	log.I.F("/event request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/events", "")
}

// EventsHandler handles a events retrieval request.
func (s *Server) EventsHandler(h handler.H) {
	log.I.F("/events request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/filter", "")
}

// FilterHandler handles a filter search request.
func (s *Server) FilterHandler(h handler.H) {
	log.I.F("/filter request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/fulltext", "")
}

// FulltextHandler handles a fulltext search request.
func (s *Server) FulltextHandler(h handler.H) {
	log.I.F("/fulltext request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/relay", "")
}

// RelayHandler handles a relay forwarding (no storage) request.
func (s *Server) RelayHandler(h handler.H) {
	log.I.F("/relay request from %s", h.RealRemote())

}

func init() {
	handler.RegisterCapability("/subscribe", "")
}

// SubscribeHandler handles a subscription (receiving new events as they come in) request.
func (s *Server) SubscribeHandler(h handler.H) {
	log.I.F("/subscribe request from %s", h.RealRemote())

}
