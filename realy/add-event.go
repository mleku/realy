package realy

import (
	"net/http"
	"regexp"
	"strings"

	"realy.lol/event"
	"realy.lol/normalize"
	"realy.lol/relay"
	eventstore "realy.lol/store"
)

var nip20prefixmatcher = regexp.MustCompile(`^\w+: `)

// AddEvent has a business rule to add an event to the relayer
func AddEvent(c Ctx, rl relay.I, ev *event.T, hr *http.Request, authedPubkey B) (accepted bool,
	message B) {
	if ev == nil {
		return false, normalize.Invalid.F("empty event")
	}

	store := rl.Storage(c)
	wrapper := &eventstore.RelayWrapper{I: store}
	advancedSaver, _ := store.(relay.AdvancedSaver)

	if !rl.AcceptEvent(c, ev, hr, authedPubkey) {
		return false, normalize.Blocked.F("event rejected by relay")
	}

	if ev.Kind.IsEphemeral() {
		// do not store ephemeral events
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(c, ev)
		}

		if saveErr := wrapper.Publish(c, ev); chk.E(saveErr) {
			switch saveErr {
			case eventstore.ErrDupEvent:
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
