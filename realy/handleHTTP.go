package realy

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"realy.lol/httpauth"
)

func (s *Server) authAdmin(r *http.Request, tolerance ...time.Duration) (authed bool,
	pubkey []byte) {
	var valid bool
	var err error
	var tolerate time.Duration
	if len(tolerance) > 0 {
		tolerate = tolerance[0]
	}
	if valid, pubkey, err = httpauth.CheckAuth(r, tolerate); chk.E(err) {
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

func (s *Server) unauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	fmt.Fprintf(w,
		"not authorized, either you did not provide an auth token or what you provided does not grant access\n")
}

func (s *Server) HandleHTTP(w http.ResponseWriter, r *http.Request) {

	log.T.S(r.Header)
	Route(w, r, Paths{
		"application/nostr+json": {
			"/relayinfo": s.handleRelayInfo,
			// methods that may need auth depending on configuration
			// "/event":  s.handleSimpleEvent,
			// "/events": s.handleEvents,
			// admin methods that require REALY_ADMIN_NPUBS auth
			"/nuke":     s.handleNuke, // todo: need some kind of confirmation scheme on this endpoint, particularly
			"/export":   s.exportHandler,
			"/import":   s.importHandler,
			"/shutdown": s.shutdownHandler,
			"":          s.defaultHandler,
		},
	})
}
