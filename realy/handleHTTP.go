package realy

import (
	"bytes"
	"fmt"
	"net/http"

	"realy.lol/httpauth"
	"realy.lol/realy/handler"
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

func (s *Server) HandleHTTP(h handler.H) {
	log.T.S(h.Request.Header)
	handler.Route(h, handler.Paths{
		"application/nostr+json": {
			"/relayinfo":    s.handleRelayInfo,
			"/capabilities": s.CapabilitiesHandler,
			"/event":        s.EventHandler,
			"/events":       s.EventsHandler,
			"/filter":       s.FilterHandler,
			"/fulltext":     s.FulltextHandler,
			"/relay":        s.RelayHandler,
			"/subscribe":    s.SubscribeHandler,
			// todo: we will use nostr+json as the codec switch for the simplified nostr
			//       http/ws on non-root paths
		},
		"": {
			"/export":   s.exportHandler,
			"/import":   s.importHandler,
			"/shutdown": s.shutdownHandler,
			"":          s.defaultHandler,
		},
	})
}
