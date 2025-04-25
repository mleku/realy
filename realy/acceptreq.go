package realy

import (
	"bytes"
	"net/http"

	"realy.mleku.dev/context"
	"realy.mleku.dev/ec/schnorr"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/log"
)

func (s *Server) AcceptReq(c context.T, hr *http.Request, id []byte,
	ff *filters.T, authedPubkey []byte, remote string) (allowed *filters.T, ok bool,
	modified bool) {

	log.T.F("%s AcceptReq pubkey %0x", remote, authedPubkey)
	authRequired := s.AuthRequired()
	if s.PublicReadable() && !authRequired {
		log.W.F("%s accept req because public readable and not auth required", remote)
		allowed = ff
		ok = true
		return
	}
	if len(s.Owners()) == 0 && !authRequired {
		log.W.F("%s accept req because no access control is enabled", remote)
		allowed = ff
		ok = true
		return
	}
	allowed = ff
	// client is permitted, pass through the filter so request/count processing does
	// not need logic and can just use the returned filter.
	// check that the client is authed to a pubkey in the owner follow list
	if len(s.Owners()) > 0 {
		s.Lock()
		for pk := range s.followed {
			if bytes.Equal(authedPubkey, []byte(pk)) {
				ok = true
				s.Unlock()
				return
			}
		}
		s.Unlock()
		// if the authed pubkey was not found, reject the request.
		log.W.F("%s reject req because %0x not on owner follow list", remote, authedPubkey)
		return
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	ok = len(authedPubkey) == schnorr.PubKeyBytesLen
	if !ok {
		log.W.F("%s reject req because auth required but user not authed", remote)
	}
	return
}
