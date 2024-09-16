package relay

import "mleku.dev/event"

func BroadcastEvent(evt *event.T) {
	notifyListeners(evt)
}
