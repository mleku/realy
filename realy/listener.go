package realy

import (
	"sync"

	"realy.lol/envelopes/eventenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/web"
)

type Listener struct {
	filters *filters.T
}

var (
	listeners      = make(map[*web.Socket]map[S]*Listener)
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

func setListener(id S, ws *web.Socket, ff *filters.T) {
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
func removeListenerId(ws *web.Socket, id S) {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	if subs, ok := listeners[ws]; ok {
		delete(listeners[ws], id)
		if len(subs) == 0 {
			delete(listeners, ws)
		}
	}
}

// Remove T conn from listeners
func removeListener(ws *web.Socket) {
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
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(id, ev); chk.E(err) {
				return
			}
			if err = res.Write(ws); chk.E(err) {
				return
			}
		}
	}
}
