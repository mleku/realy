package realy

import "realy.lol/event"

func BroadcastEvent(evt *event.T) {
	notifyListeners(evt)
}
