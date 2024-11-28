package realy

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"realy.lol/event"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/relay/wrapper"
	"realy.lol/store"
)

var nip20prefixmatcher = regexp.MustCompile(`^\w+: `)

// AddEvent has a business rule to add an event to the event store
func AddEvent(c Ctx, rl relay.I, ev *event.T, hr *http.Request, origin S,
	authedPubkey B) (accepted bool, message B) {
	if ev == nil {
		return false, normalize.Invalid.F("empty event")
	}

	sto := rl.Storage(c)
	wrapper := &wrapper.RelayWrapper{I: sto}
	advancedSaver, _ := sto.(relay.AdvancedSaver)

	accept, notice := rl.AcceptEvent(c, ev, hr, origin, authedPubkey)
	if !accept {
		return false, normalize.Blocked.F(notice)
	}
	if ev.Tags.ContainsProtectedMarker() {
		if len(authedPubkey) == 0 || !equals(ev.PubKey, authedPubkey) {
			return false, B(fmt.Sprintf(
				"event with relay marker tag '-' may only be published by matching npub: %0x is not %0x",
				authedPubkey, ev.PubKey))
		}
	}
	if ev.Kind.IsEphemeral() {
		// do not store ephemeral events
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(c, ev)
		}

		if saveErr := wrapper.Publish(c, ev); chk.E(saveErr) {
			switch saveErr {
			case store.ErrDupEvent:
				return false, normalize.Error.F(saveErr.Error())
			default:
				errmsg := saveErr.Error()
				if nip20prefixmatcher.MatchString(errmsg) {
					if strings.Contains(errmsg, "tombstone") {
						return false, normalize.Blocked.F(
							"event was deleted, not storing it again")
					}
					return false, normalize.Error.F(errmsg)
				} else {
					return false, normalize.Error.F("failed to save (%s)", errmsg)
				}
			}
			// } else {
			// 	log.D.F("saved event %s", ev.Serialize())
		}

		if advancedSaver != nil {
			advancedSaver.AfterSave(ev)
		}
	}

	var authRequired bool
	if ar, ok := rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	notifyListeners(authRequired, ev)

	accepted = true
	return
}
