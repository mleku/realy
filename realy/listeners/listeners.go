// Package listeners is a singleton package that keeps track of subscriptions in
// both websockets and http SSE, including managing the authentication state of
// a connection.
package listeners

import (
	"bytes"
	"crypto/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/fasthttp/websocket"

	"realy.lol/bech32encoding"
	"realy.lol/context"
	"realy.lol/ec/bech32"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/tag"
	"realy.lol/units"
	"realy.lol/web"
)

type (
	L struct{ filters *filters.T }

	Map map[*web.Socket]map[string]*L

	H struct {
		// Ctx is the http.Request context of the subscriber, this enables garbage
		// collecting the subscriptions from http.
		Ctx context.T
		// Receiver is a channel that the listener sends subscription events to for http
		// subscribe endpoint.
		Receiver event.C
		// Pubkey is the pubkey authed to this subscription
		Pubkey []byte
		// Filter is the filter associated with the http subscription
		Filter *filter.T
	}

	Subs map[*H]struct{}

	T struct {
		Ctx context.T
		sync.Mutex
		Map
		Subs
		Hchan        chan H
		Hlock        sync.Mutex
		ChallengeHRP string
		WriteWait,
		PongWait,
		PingPeriod time.Duration
		MaxMessageSize  int64
		ChallengeLength int
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
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
)

func New(ctx context.T) (l *T) {
	l = &T{
		Ctx:             ctx,
		Map:             make(Map),
		Subs:            make(Subs),
		Hchan:           make(chan H),
		ChallengeHRP:    DefaultChallengeHRP,
		WriteWait:       DefaultWriteWait,
		PongWait:        DefaultPongWait,
		PingPeriod:      DefaultPingPeriod,
		MaxMessageSize:  DefaultMaxMessageSize,
		ChallengeLength: DefaultChallengeLength,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case h := <-l.Hchan:
				l.Hlock.Lock()
				l.Subs[&h] = struct{}{}
				l.Hlock.Unlock()
			}
		}
	}()
	return
}

func (l *T) GetChallenge(conn *websocket.Conn, req *http.Request,
	addr string) (ws *web.Socket) {
	var err error
	cb := make([]byte, l.ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		panic(err)
	}
	var b5 []byte
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded []byte
	if encoded, err = bech32.Encode([]byte(l.ChallengeHRP), b5); chk.E(err) {
		return
	}
	ws = web.NewSocket(conn, req, encoded)
	return
}

func (l *T) SetListener(id string, ws *web.Socket, ff *filters.T) {
	l.Mutex.Lock()
	subs, ok := l.Map[ws]
	if !ok {
		subs = make(map[string]*L)
		l.Map[ws] = subs
	}
	subs[id] = &L{filters: ff}
	l.Mutex.Unlock()
}

func (l *T) RemoveListenerId(ws *web.Socket, id string) {
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

func (l *T) NotifyListeners(authRequired, publicReadable bool, ev *event.T) {
	if ev == nil {
		return
	}
	var err error
	l.Mutex.Lock()
	for ws, subs := range l.Map {
		for id, listener := range subs {
			if !publicReadable {
				if authRequired && !ws.IsAuthed() {
					continue
				}
			}
			if !listener.filters.Match(ev) {
				continue
			}
			if ev.Kind.IsPrivileged() {
				ab := ws.AuthedBytes()
				var containsPubkey bool
				if ev.Tags != nil {
					containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(ab))
				}
				if !bytes.Equal(ev.Pubkey, ab) || containsPubkey {
					log.I.F("authed user %0x not privileged to receive event\n%s",
						ab, ev.Serialize())
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
	l.Mutex.Unlock()
	l.Hlock.Lock()
	var subs []*H
	for sub := range l.Subs {
		// check if the subscription's subscriber is still alive
		select {
		case <-sub.Ctx.Done():
			subs = append(subs, sub)
		default:
		}
	}
	for _, sub := range subs {
		delete(l.Subs, sub)
	}
	subs = subs[:0]
	for sub := range l.Subs {
		// if auth required, check the subscription pubkey matches
		if !publicReadable {
			if authRequired && len(sub.Pubkey) == 0 {
				continue
			}
		}
		// if the filter doesn't match, skip
		if !sub.Filter.Matches(ev) {
			continue
		}
		// if the filter is privileged and the user doesn't have matching auth, skip
		if ev.Kind.IsPrivileged() {
			ab := sub.Pubkey
			var containsPubkey bool
			if ev.Tags != nil {
				containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(ab))
			}
			if !bytes.Equal(ev.Pubkey, ab) || containsPubkey {
				continue
			}
		}
		// send the event to the subscriber
		sub.Receiver <- ev
	}
	l.Hlock.Unlock()
}
