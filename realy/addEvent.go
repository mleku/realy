package realy

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/log"
	"realy.mleku.dev/normalize"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/socketapi"
	"realy.mleku.dev/store"
)

func (s *Server) addEvent(c context.T, rl relay.I, ev *event.T,
	hr *http.Request, origin string,
	authedPubkey []byte) (accepted bool, message []byte) {

	if ev == nil {
		return false, normalize.Invalid.F("empty event")
	}
	sto := rl.Storage()
	advancedSaver, _ := sto.(relay.AdvancedSaver)
	// don't allow storing event with protected marker as per nip-70 with auth enabled.
	if (s.authRequired || !s.publicReadable) && ev.Tags.ContainsProtectedMarker() {
		if len(authedPubkey) == 0 || !bytes.Equal(ev.Pubkey, authedPubkey) {
			return false,
				[]byte(fmt.Sprintf("event with relay marker tag '-' (nip-70 protected event) "+
					"may only be published by matching npub: %0x is not %0x",
					authedPubkey, ev.Pubkey))
		}
	}
	if ev.Kind.IsEphemeral() {
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(c, ev)
		}
		if saveErr := s.Publish(c, ev); saveErr != nil {
			if errors.Is(saveErr, store.ErrDupEvent) {
				return false, normalize.Error.F(saveErr.Error())
			}
			errmsg := saveErr.Error()
			if socketapi.NIP20prefixmatcher.MatchString(errmsg) {
				if strings.Contains(errmsg, "tombstone") {
					return false, normalize.Blocked.F("event was deleted, not storing it again")
				}
				if strings.HasPrefix(errmsg, string(normalize.Blocked)) {
					return false, []byte(errmsg)
				}
				return false, normalize.Error.F(errmsg)
			} else {
				return false, normalize.Error.F("failed to save (%s)", errmsg)
			}
		}
		if advancedSaver != nil {
			advancedSaver.AfterSave(ev)
		}
	}
	var authRequired bool
	if ar, ok := rl.(relay.Authenticator); ok {
		authRequired = ar.AuthRequired()
	}
	// notify subscribers
	s.listeners.Deliver(authRequired, s.publicReadable, ev)
	accepted = true
	log.I.F("event id %0x stored", ev.Id)
	return
}
