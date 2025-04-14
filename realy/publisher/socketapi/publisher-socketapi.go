package socketapi

import (
	"bytes"
	"regexp"
	"sync"
	"time"

	"realy.mleku.dev/context"
	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/units"
	"realy.mleku.dev/ws"
)

var (
	NIP20prefixmatcher = regexp.MustCompile(`^\w+: `)
)

const (
	DefaultWriteWait      = 10 * time.Second
	DefaultPongWait       = 60 * time.Second
	DefaultPingPeriod     = DefaultPongWait / 2
	DefaultMaxMessageSize = 1 * units.Mb
)

// Map is a map of filters associated with a collection of ws.Listener connections.
type Map map[*ws.Listener]map[string]*filters.T

type W struct {
	*ws.Listener
	// If Cancel is true, this is a close command.
	Cancel bool
	// Id is the subscription Id. If Cancel is true, cancel the named subscription, otherwise,
	// cancel the publisher for the socket.
	Id       string
	Receiver event.C
	Filters  *filters.T
}

type Close struct {
	*ws.Listener
	Id string
}

type WSP struct {
	// Chan is a channel that socekt api subscriptions send their receiver channel through.
	Chan chan W
	// Mx is the mutex for the Map.
	Mx sync.Mutex
	// Map is the map of subscribers and subscriptions from the websocket api.
	Map
	// WsPingWait is the time between writing pings to the websocket.
	WsPingWait time.Duration
	// WsPongWait is the time after which the connection will be considered ded.
	WsPongWait time.Duration
	// WsPingPeriod sets the time between sending pings to the client.
	WsPingPeriod time.Duration
	// WsMaxMessageSize is the largest message that will be allowed to be received on the
	// websocket.
	WsMaxMessageSize int64
}

func NewWSP() *WSP {
	return &WSP{
		Chan:             make(chan W, 32),
		Map:              make(Map),
		WsPingWait:       DefaultWriteWait,
		WsPongWait:       DefaultPongWait,
		WsPingPeriod:     DefaultPingPeriod,
		WsMaxMessageSize: DefaultMaxMessageSize,
	}
}

func (wsp *WSP) ReceiverLoop(ctx context.T) {
	for {
		select {
		case <-ctx.Done():
			return
		case h := <-wsp.Chan:
			if h.Cancel {
				if h.Id == "" {
					wsp.removeSubscriber(h.Listener)
				} else {
					wsp.removeSubscriberId(h.Listener, h.Id)
				}
				continue
			}
			wsp.Mx.Lock()
			subs, ok := wsp.Map[h.Listener]
			if !ok {
				subs = make(map[string]*filters.T)
				wsp.Map[h.Listener] = subs
			}
			subs[h.Id] = h.Filters
			wsp.Mx.Unlock()
		}
	}
}

// removeSubscriberId removes a specific subscription from a subscriber websocket.
func (wsp *WSP) removeSubscriberId(ws *ws.Listener, id string) {
	wsp.Mx.Lock()
	var subs map[string]*filters.T
	var ok bool
	if subs, ok = wsp.Map[ws]; ok {
		delete(wsp.Map[ws], id)
		_ = subs
		if len(subs) == 0 {
			delete(wsp.Map, ws)
		}
	}
	wsp.Mx.Unlock()
}

// removeSubscriber removes a websocket from the S collection.
func (wsp *WSP) removeSubscriber(ws *ws.Listener) {
	wsp.Mx.Lock()
	clear(wsp.Map[ws])
	delete(wsp.Map, ws)
	wsp.Mx.Unlock()
}

func (wsp *WSP) Deliver(authRequired, publicReadable bool, ev *event.T) {
	var err error
	wsp.Mx.Lock()
	for ws, subs := range wsp.Map {
		for id, subscriber := range subs {
			if !publicReadable {
				if authRequired && !ws.IsAuthed() {
					continue
				}
			}
			if !subscriber.Match(ev) {
				continue
			}
			if ev.Kind.IsPrivileged() {
				ab := ws.AuthedBytes()
				var containsPubkey bool
				if ev.Tags != nil {
					containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(ab))
				}
				if !bytes.Equal(ev.Pubkey, ab) || containsPubkey {
					if ab == nil {
						continue
					}
					continue
				}
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(id, ev); chk.E(err) {
				continue
			}
			if err = res.Write(ws); chk.E(err) {
				continue
			}
		}
	}
	wsp.Mx.Unlock()

}
