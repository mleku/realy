// Package listeners is a singleton package that keeps track of nostr websockets
package listeners

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
	L   struct{ filters *filters.T }
	Map map[*web.Socket]map[st]*L
	T   struct {
		sync.Mutex
		Map
		ChallengeHRP st
		WriteWait,
		PongWait,
		PingPeriod time.Duration
		MaxMessageSize  int64
		ChallengeLength no
	}
)

const (
	DefaultChallengeHRP    = "nchal"
	DefaultWriteWait       = 10 * time.Second
	DefaultPongWait        = 60 * time.Second
	DefaultPingPeriod      = DefaultPongWait / 2
	DefaultMaxMessageSize  = 1 * units.Mb
	DefaultChallengeLength = 16
)

var (
	NIP20prefixmatcher = regexp.MustCompile(`^\w+: `)
	Upgrader           = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bo {
			return true
		}}
)

func New() (l *T) {
	return &T{
		Map:             make(Map),
		ChallengeHRP:    DefaultChallengeHRP,
		WriteWait:       DefaultWriteWait,
		PongWait:        DefaultPongWait,
		PingPeriod:      DefaultPingPeriod,
		MaxMessageSize:  DefaultMaxMessageSize,
		ChallengeLength: DefaultChallengeLength,
	}
}

func (l *T) GetChallenge(conn *websocket.Conn, req *http.Request,
	addr string) (ws *web.Socket) {
	var err er
	cb := make(by, l.ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		panic(err)
	}
	var b5 by
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded by
	if encoded, err = bech32.Encode(by(l.ChallengeHRP), b5); chk.E(err) {
		return
	}
	ws = web.NewSocket(conn, req, encoded)
	return
}

func (l *T) SetListener(id st, ws *web.Socket, ff *filters.T) {
	l.Mutex.Lock()
	subs, ok := l.Map[ws]
	if !ok {
		subs = make(map[st]*L)
		l.Map[ws] = subs
	}
	subs[id] = &L{filters: ff}
	l.Mutex.Unlock()
}

func (l *T) RemoveListenerId(ws *web.Socket, id st) {
	l.Mutex.Lock()
	if subs, ok := l.Map[ws]; ok {
		delete(l.Map[ws], id)
		if len(subs) == 0 {
			delete(l.Map, ws)
		}
	}
	l.Mutex.Unlock()
}

func (l *T) RemoveListener(ws *web.Socket) {
	l.Mutex.Lock()
	clear(l.Map[ws])
	delete(l.Map, ws)
	l.Mutex.Unlock()
}

func (l *T) NotifyListeners(authRequired bo, ev *event.T) {
	if ev == nil {
		return
	}
	var err er
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	for ws, subs := range l.Map {
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
