package realy

import (
	"bytes"
	"fmt"
	"net/http"

	"realy.lol/httpauth"
	"realy.lol/realy/api"
	"realy.lol/realy/router"
)

func (s *Server) auth(r *http.Request) (authed bool) {
	var valid bool
	var pubkey []byte
	var err error
	if valid, pubkey, err = httpauth.ValidateRequest(r); chk.E(err) {
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

func (s *Server) HandleHTTP(h api.H) {
	log.T.S(h.Request.Header)
	router.Route(h, "")
}
