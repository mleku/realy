package realy

import (
	"bytes"
	"fmt"
	"net/http"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/tag"
)

func (s *Server) JWTVerifyFunc(npub string) (jwtPub string, pk []byte, err error) {
	if pk, err = hex.Dec(npub); chk.E(err) {
		return
	}
	var evs event.Ts
	evs, err = s.relay.Storage().QueryEvents(context.Bg(), &filter.T{
		Authors: tag.New(pk), Kinds: kinds.New(kind.JWTBinding),
	})
	if chk.E(err) {
		return
	}
	// there should only be one
	if len(evs) < 1 {
		err = errorf.E("JWT Binding event not found for pubkey %s", npub)
		return
	}
	// we will only use the first one, the query should only return the newest per
	// pubkey/kind for kind.JWTBinding as it is a replaceable event.
	ev := evs[0]
	jtag := ev.Tags.GetAll(tag.New("J"))
	if jtag.Len() < 1 {
		err = errorf.E("JWT Binding event tag not found for pubkey %s\n%s\n",
			npub, ev.SerializeIndented())
		return
	}
	jwtPub = string(jtag.F()[0].Value())
	return
}

func (s *Server) authAdmin(r *http.Request) (authed bool) {
	var valid bool
	var pubkey []byte
	var err error
	if valid, pubkey, err = httpauth.CheckAuth(r, s.JWTVerifyFunc); chk.E(err) {
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
			"/event":  s.handleSimpleEvent,
			"/events": s.handleEvents,
			// admin methods that require REALY_ADMIN_NPUBS auth
			"/nuke":     s.handleNuke, // todo: need some kind of confirmation scheme on this endpoint, particularly
			"/export":   s.exportHandler,
			"/import":   s.importHandler,
			"/shutdown": s.shutdownHandler,
			"":          s.defaultHandler,
		},
	})
}
