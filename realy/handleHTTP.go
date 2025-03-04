package realy

import (
	"bytes"
	"fmt"
	"net/http"

	"realy.lol/httpauth"
)

func (s *Server) auth(r *http.Request) (authed bool) {
	var valid bool
	var pubkey []byte
	var err error
	// todo: need to add the verifier function for JWT
	if valid, pubkey, err = httpauth.ValidateRequest(r, nil); chk.E(err) {
		return
	}
	if !valid {
		return
	}
	// check admins pubkey list
	for _, v := range s.admins {
		if bytes.Equal(v.Pub(), pubkey) {
			authed = true
			return
		}
	}
	return
}

func (s *Server) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	fmt.Fprintf(w, "your npub is not welcome here\n")
}

func (s *Server) HandleHTTP(h Handler) {
	log.T.S(h.Request.Header)
	Route(h, Paths{
		"application/nostr+json": {
			"/relayinfo": s.handleRelayInfo,
			"/event":     s.handleSimpleEvent,
		},
		"": {
			"/export":   s.exportHandler,
			"/import":   s.importHandler,
			"/shutdown": s.shutdownHandler,
			"":          s.defaultHandler,
		},
	})
}
