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
	if s.PublicReadable() && !s.AuthRequired() {
		log.T.F("%s accept because public readable and not auth required", remote)
		allowed = ff
		ok = true

	}
	if len(s.Owners()) == 0 && !s.AuthRequired() {
		log.T.F("%s accept because no access control is enabled", remote)
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
		return
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	ok = len(authedPubkey) == schnorr.PubKeyBytesLen
	return
}
