package realy

import (
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
		log.I.F("import of event data requested on admin port", r.RequestURI)
		store := s.relay.Storage(context.Bg())
		read := io.LimitReader(r.Body, r.ContentLength)
		store.Import(read)
		w.Write(B("ok"))
	}
}
