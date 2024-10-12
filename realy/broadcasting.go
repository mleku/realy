package realy

import "realy.lol/event"

func BroadcastEvent(authRequired bool, ev *event.T) {
	notifyListeners(authRequired, ev)
}
