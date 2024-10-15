package realy

import (
	"net/http"

	"realy.lol/context"
)

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/export":
		log.I.F("export of event data requested on admin port")
		store := s.relay.Storage(context.Bg())
		store.Export(w)
	}
}
