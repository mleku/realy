package realy

import (
	"fmt"
	"io"
	"net/http"

	"realy.lol/context"
)

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	log.I.S(r.Header)
	switch r.URL.Path {
	case "/export":
		log.I.F("export of event data requested on admin port")
		store := s.relay.Storage(context.Bg())
		store.Export(w)
	case "/import":
		log.I.F("import of event data requested on admin port %s", r.RequestURI)
		store := s.relay.Storage(context.Bg())
		read := io.LimitReader(r.Body, r.ContentLength)
		store.Import(read)
	case "/shutdown":
		fmt.Fprintf(w, "shutting down")
		defer r.Body.Close()
		s.Shutdown(context.Bg())
	}
}
