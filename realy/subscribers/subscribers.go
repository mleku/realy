// Package subscribers is a singleton package that keeps track of subscriptions in
// both websockets and http SSE, including managing the authentication state of
// a connection.
package subscribers

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
	"realy.lol/ws"
)

type (
	// L is a collection of filters.
	L struct{ filters *filters.T }

	// Map is a map of filters associated with a collection of web.Listener..
	Map map[*ws.Listener]map[string]*L

	// H is the control structure for a HTTP SSE subscription, including the filter, authed
	// pubkey and a channel to send the events to.
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

	// Subs is a collection of H TTP subscriptions.
	Subs map[*H]struct{}

	// S is the control structure for the subscription management scheme.
	S struct {
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

// New creates a new subscribers.S.
func New(ctx context.T) (l *S) {
	l = &S{
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

// GetChallenge generates a new challenge for a subscriber.
func (s *S) GetChallenge(conn *websocket.Conn, req *http.Request,
	addr string) (w *ws.Listener) {
	var err error
	cb := make([]byte, s.ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		panic(err)
	}
	var b5 []byte
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded []byte
	if encoded, err = bech32.Encode([]byte(s.ChallengeHRP), b5); chk.E(err) {
		return
	}
	w = ws.NewListener(conn, req, encoded)
	return
}

// Set a new subscriber with its collection of filters.
func (s *S) Set(id string, ws *ws.Listener, ff *filters.T) {
	s.Mutex.Lock()
	subs, ok := s.Map[ws]
	if !ok {
		subs = make(map[string]*L)
		s.Map[ws] = subs
	}
	subs[id] = &L{filters: ff}
	s.Mutex.Unlock()
}

// RemoveSubscriberId removes a specific subscription from a subscriber websocket.
func (s *S) RemoveSubscriberId(ws *ws.Listener, id string) {
	s.Mutex.Lock()
	if subs, ok := s.Map[ws]; ok {
		delete(s.Map[ws], id)
		if len(subs) == 0 {
			delete(s.Map, ws)
		}
	}
	s.Mutex.Unlock()
}

// RemoveSubscriber removes a websocket from the subscribers.S collection.
func (s *S) RemoveSubscriber(ws *ws.Listener) {
	s.Mutex.Lock()
	clear(s.Map[ws])
	delete(s.Map, ws)
	s.Mutex.Unlock()
}

// NotifySubscribers processes a new event and determines whether to send it to subscribers.
func (s *S) NotifySubscribers(authRequired, publicReadable bool, ev *event.T) {
	if ev == nil {
		return
	}
	var err error
	s.Mutex.Lock()
	for ws, subs := range s.Map {
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
					if ab == nil {
						continue
					}
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
	s.Mutex.Unlock()
	s.Hlock.Lock()
	var subs []*H
	for sub := range s.Subs {
		// check if the subscription's subscriber is still alive
		select {
		case <-sub.Ctx.Done():
			subs = append(subs, sub)
		default:
		}
	}
	for _, sub := range subs {
		delete(s.Subs, sub)
	}
	subs = subs[:0]
	for sub := range s.Subs {
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
	s.Hlock.Unlock()
}
