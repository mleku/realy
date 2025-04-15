package socketapi

import (
	"bytes"
	"regexp"
	"sync"

	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/realy/publish/publisher"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/ws"
)

const Type = "socketapi"

var (
	NIP20prefixmatcher = regexp.MustCompile(`^\w+: `)
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

func (w *W) Type() string { return Type }

type Close struct {
	*ws.Listener
	Id string
}

type S struct {
	// Mx is the mutex for the Map.
	Mx sync.Mutex
	// Map is the map of subscribers and subscriptions from the websocket api.
	Map
}

var _ publisher.I = &S{}

func New() *S { return &S{Map: make(Map)} }

func (p *S) Type() string { return Type }

func (p *S) Receive(msg publisher.Message) {
	if m, ok := msg.(*W); ok {
		if m.Cancel {
			if m.Id == "" {
				p.removeSubscriber(m.Listener)
			} else {
				p.removeSubscriberId(m.Listener, m.Id)
			}
			return
		}
		p.Mx.Lock()
		if subs, ok := p.Map[m.Listener]; !ok {
			subs = make(map[string]*filters.T)
			p.Map[m.Listener] = subs
		} else {
			subs[m.Id] = m.Filters
		}
		p.Mx.Unlock()

	}
}

func (p *S) Deliver(authRequired, publicReadable bool, ev *event.T) {
	var err error
	p.Mx.Lock()
	for w, subs := range p.Map {
		for id, subscriber := range subs {
			if !publicReadable {
				if authRequired && !w.IsAuthed() {
					continue
				}
			}
			if !subscriber.Match(ev) {
				continue
			}
			if ev.Kind.IsPrivileged() {
				ab := w.AuthedBytes()
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
			if err = res.Write(w); chk.E(err) {
				continue
			}
		}
	}
	p.Mx.Unlock()
}

// removeSubscriberId removes a specific subscription from a subscriber websocket.
func (p *S) removeSubscriberId(ws *ws.Listener, id string) {
	p.Mx.Lock()
	var subs map[string]*filters.T
	var ok bool
	if subs, ok = p.Map[ws]; ok {
		delete(p.Map[ws], id)
		_ = subs
		if len(subs) == 0 {
			delete(p.Map, ws)
		}
	}
	p.Mx.Unlock()
}

// removeSubscriber removes a websocket from the S collection.
func (p *S) removeSubscriber(ws *ws.Listener) {
	p.Mx.Lock()
	clear(p.Map[ws])
	delete(p.Map, ws)
	p.Mx.Unlock()
}
