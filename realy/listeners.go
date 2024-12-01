package realy

import (
	"crypto/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/fasthttp/websocket"

	"realy.lol/bech32encoding"
	"realy.lol/ec/bech32"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/event"
	"realy.lol/filters"
	"realy.lol/tag"
	"realy.lol/units"
	"realy.lol/web"
)

type (
	Listener struct{ filters *filters.T }
)

const (
	ChallengeHRP    = "nchal"
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = pongWait / 2
	maxMessageSize  = 1 * units.Mb
	ChallengeLength = 16
)

var (
	nip20prefixmatcher = regexp.MustCompile(`^\w+: `)
	upgrader           = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bo {
			return true
		}}
	listeners      = make(map[*web.Socket]map[st]*Listener)
	listenersMutex sync.Mutex
)

func challenge(conn *websocket.Conn, req *http.Request, addr string) (ws *web.Socket) {
	var err er
	cb := make(by, ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		panic(err)
	}
	var b5 by
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded by
	if encoded, err = bech32.Encode(by(ChallengeHRP), b5); chk.E(err) {
		return
	}
	ws = web.NewSocket(conn, req, encoded)
	return
}

func setListener(id st, ws *web.Socket, ff *filters.T) {
	listenersMutex.Lock()
	subs, ok := listeners[ws]
	if !ok {
		subs = make(map[st]*Listener)
		listeners[ws] = subs
	}
	subs[id] = &Listener{filters: ff}
	listenersMutex.Unlock()
}

func removeListenerId(ws *web.Socket, id st) {
	listenersMutex.Lock()
	if subs, ok := listeners[ws]; ok {
		delete(listeners[ws], id)
		if len(subs) == 0 {
			delete(listeners, ws)
		}
	}
	listenersMutex.Unlock()
}

func removeListener(ws *web.Socket) {
	listenersMutex.Lock()
	clear(listeners[ws])
	delete(listeners, ws)
	listenersMutex.Unlock()
}

func notifyListeners(authRequired bo, ev *event.T) {
	if ev == nil {
		return
	}
	var err er
	listenersMutex.Lock()
	defer listenersMutex.Unlock()
	for ws, subs := range listeners {
		for id, listener := range subs {
			if authRequired && !ws.IsAuthed() {
				continue
			}
			if !listener.filters.Match(ev) {
				continue
			}
			if ev.Kind.IsPrivileged() {
				ab := ws.AuthedBytes()
				var containsPubkey bo
				if ev.Tags != nil {
					containsPubkey = ev.Tags.ContainsAny(by{'p'}, tag.New(ab))
				}
				if !equals(ev.PubKey, ab) || containsPubkey {
					log.I.F("authed user %0x not privileged to receive event\n%s",
						ws.AuthedBytes(), ev.Serialize())
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
}
