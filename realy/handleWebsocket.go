package realy

import (
	"net/http"

	"realy.mleku.dev/socketapi"
)

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	a := &socketapi.A{Server: s, ClientsMu: &s.clientsMu, Clients: s.clients}
	a.Serve(w, r, s)
}
