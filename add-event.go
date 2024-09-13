package realy

import (
	"fmt"
	"regexp"

	. "nostr.mleku.dev"
	"nostr.mleku.dev/codec/event"
	"store.mleku.dev"
)

var nip20prefixmatcher = regexp.MustCompile(`^\w+: `)

// AddEvent has a business rule to add an event to the relayer
func AddEvent(c Ctx, relay Relay, evt *event.T) (accepted bool, message S) {
	if evt == nil {
		return false, ""
	}

	store := relay.Storage(c)
	wrapper := &eventstore.RelayWrapper{store}
	advancedSaver, _ := store.(AdvancedSaver)

	if !relay.AcceptEvent(c, evt) {
		return false, "blocked: event blocked by relay"
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
				return true, saveErr.Error()
			default:
				errmsg := saveErr.Error()
				if nip20prefixmatcher.MatchString(errmsg) {
					return false, errmsg
				} else {
					return false, fmt.Sprintf("error: failed to save (%s)", errmsg)
				}
			}
		}

		if advancedSaver != nil {
			advancedSaver.AfterSave(evt)
		}
	}

	notifyListeners(evt)

	return true, ""
}
