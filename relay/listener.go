package relay

import (
	"sync"

	"mleku.dev/envelopes/eventenvelope"
	"mleku.dev/event"
	"mleku.dev/filter"
	"mleku.dev/filters"
)

type Listener struct {
	filters *filters.T
}

var (
	listeners      = make(map[*WebSocket]map[S]*Listener)
	listenersMutex sync.Mutex
)

func GetListeningFilters() *filters.T {
	respfilters := &filters.T{F: make([]*filter.T, 0, len(listeners)*2)}

	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	// here we go through all the existing listeners
	for _, connlisteners := range listeners {
		for _, listener := range connlisteners {
			for _, listenerfilter := range listener.filters.F {
				for _, respfilter := range respfilters.F {
					// check if this filter specifically is already added to respfilters
					if filter.Equal(listenerfilter, respfilter) {
						goto nextconn
					}
				}

				// field not yet present on respfilters, add it
				respfilters.F = append(respfilters.F, listenerfilter)

				// continue to the next filter
			nextconn:
				continue
			}
		}
	}

	// respfilters will be a slice with all the distinct filter we currently have active
	return respfilters
}

func setListener(id S, ws *WebSocket, ff *filters.T) {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	subs, ok := listeners[ws]
	if !ok {
		subs = make(map[S]*Listener)
		listeners[ws] = subs
	}

	subs[id] = &Listener{filters: ff}
}

// Remove a specific subscription id from listeners for a given ws client
func removeListenerId(ws *WebSocket, id S) {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	if subs, ok := listeners[ws]; ok {
		delete(listeners[ws], id)
		if len(subs) == 0 {
			delete(listeners, ws)
		}
	}
}

// Remove WebSocket conn from listeners
func removeListener(ws *WebSocket) {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()
	clear(listeners[ws])
	delete(listeners, ws)
}

func notifyListeners(ev *event.T) {
	var err E
	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	for ws, subs := range listeners {
		for id, listener := range subs {
			if !listener.filters.Match(ev) {
				continue
			}
			if err = eventenvelope.NewResultWith(id, ev).Write(ws); chk.E(err) {
				return
			}
			// ws.WriteJSON(nostr.EventEnvelope{SubscriptionID: &id, Event: *ev})
		}
	}
}