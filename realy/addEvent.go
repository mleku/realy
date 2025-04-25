package realy

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/log"
	"realy.mleku.dev/publish"
	"realy.mleku.dev/reason"
	"realy.mleku.dev/store"
)

var (
	NIP20prefixmatcher = regexp.MustCompile(`^\w+: `)
)

func (s *Server) addEvent(c context.T, ev *event.T,
	authedPubkey []byte, remote string) (accepted bool, message []byte) {

	authRequired := s.AuthRequired()
	if ev == nil {
		log.I.F("empty event")
		return false, reason.Invalid.F("empty event")
	}
	// don't allow storing event with protected marker as per nip-70 with auth enabled.
	if (authRequired || !s.PublicReadable()) && ev.Tags.ContainsProtectedMarker() {
		if len(authedPubkey) == 0 || !bytes.Equal(ev.Pubkey, authedPubkey) {
			return false,
				[]byte(fmt.Sprintf("event with relay marker tag '-' (nip-70 protected event) "+
					"may only be published by matching npub: %0x is not %0x",
					authedPubkey, ev.Pubkey))
		}
	}
	if ev.Kind.IsEphemeral() {
	} else {
		if saveErr := s.Publish(c, ev); saveErr != nil {
			if errors.Is(saveErr, store.ErrDupEvent) {
				return false, reason.Error.F(saveErr.Error())
			}
			errmsg := saveErr.Error()
			if NIP20prefixmatcher.MatchString(errmsg) {
				if strings.Contains(errmsg, "tombstone") {
					return false, reason.Blocked.F("event was deleted, not storing it again")
				}
				if strings.HasPrefix(errmsg, string(reason.Blocked)) {
					return false, []byte(errmsg)
				}
				return false, reason.Error.F(errmsg)
			} else {
				return false, reason.Error.F("failed to save (%s)", errmsg)
			}
		}
	}
	// notify subscribers
	publish.P.Deliver(authRequired, s.PublicReadable(), ev)
	accepted = true
	log.T.F("event id %0x stored", ev.Id)
	return
}
