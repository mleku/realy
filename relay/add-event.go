package relay

import (
	"regexp"
	"strings"

	"realy.lol/event"
	"realy.lol/normalize"
	eventstore "realy.lol/store"
)

var nip20prefixmatcher = regexp.MustCompile(`^\w+: `)

// AddEvent has a business rule to add an event to the relayer
func AddEvent(c Ctx, relay Relay, evt *event.T) (accepted bool, message B) {
	if evt == nil {
		return false, normalize.Blocked.F("empty event")
	}

	store := relay.Storage(c)
	wrapper := &eventstore.RelayWrapper{I: store}
	advancedSaver, _ := store.(AdvancedSaver)

	if !relay.AcceptEvent(c, evt) {
		return false, normalize.Blocked.F("event blocked by relay")
	}

	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(c, evt)
		}

		if saveErr := wrapper.Publish(c, evt); saveErr != nil {
			switch saveErr {
			case eventstore.ErrDupEvent:
				return true, normalize.Error.F(saveErr.Error())
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
		}

		if advancedSaver != nil {
			advancedSaver.AfterSave(evt)
		}
	}

	notifyListeners(evt)

	accepted = true
	return
}
