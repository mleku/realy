// Package subscribers is a singleton package that keeps track of subscriptions in
// both websockets and http SSE, including managing the authentication state of
// a connection.
package subscribers

import (
	"crypto/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/fasthttp/websocket"

	"realy.mleku.dev/bech32encoding"
	"realy.mleku.dev/context"
	"realy.mleku.dev/ec/bech32"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/units"
	"realy.mleku.dev/ws"
)

type (
	// L is a collection of filters.
	L struct{ filters *filters.T }

	// WsMap is a map of filters associated with a collection of ws.Listener connections.
	WsMap map[*ws.Listener]map[string]*L

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

	// HMap is a collection of H TTP subscriptions.
	HMap map[*H]struct{}

	// S is the control structure for the subscription management scheme.
	S struct {
		Ctx context.T

		// WsMx is the mutex for the WsMap.
		WsMx sync.Mutex
		// WsMap is the map of subscribers and subscriptions from the websocket api.
		WsMap
		// WsPingWait is the time between writing pings to the websocket.
		WsPingWait time.Duration
		// WsPongWait is the time after which the connection will be considered ded.
		WsPongWait time.Duration
		// WsPingPeriod sets the time between sending pings to the client.
		WsPingPeriod time.Duration
		// WsMaxMessageSize is is the largest message that will be allowed to be received on the
		// websocket.
		WsMaxMessageSize int64

		// HMap is the map of subscriptions from the http api.
		HMap
		// HChan is a channel that http api subscriptions send their receiver channel through.
		HChan chan H
		// HLock is the mutex that locks the HMap map.
		HMx sync.Mutex
		// ChallengeHRP is the bech32 HRP prefix used in encoding challenges.
		ChallengeHRP string
		// ChallengeLength is the length of bytes of the challenge random value.
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
)

// New creates a new subscribers.S.
func New(ctx context.T) (l *S) {
	l = &S{
		Ctx:              ctx,
		WsMap:            make(WsMap),
		HMap:             make(HMap),
		HChan:            make(chan H),
		ChallengeHRP:     DefaultChallengeHRP,
		WsPingWait:       DefaultWriteWait,
		WsPongWait:       DefaultPongWait,
		WsPingPeriod:     DefaultPingPeriod,
		WsMaxMessageSize: DefaultMaxMessageSize,
		ChallengeLength:  DefaultChallengeLength,
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case h := <-l.HChan:
				l.HMx.Lock()
				l.HMap[&h] = struct{}{}
				l.HMx.Unlock()
			}
		}
	}()
	return
}

// GetChallenge generates a new challenge for a subscriber.
func (s *S) GetChallenge(conn *websocket.Conn, req *http.Request) (w *ws.Listener) {
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
	s.WsMx.Lock()
	subs, ok := s.WsMap[ws]
	if !ok {
		subs = make(map[string]*L)
		s.WsMap[ws] = subs
	}
	subs[id] = &L{filters: ff}
	s.WsMx.Unlock()
}

// RemoveSubscriberId removes a specific subscription from a subscriber websocket.
func (s *S) RemoveSubscriberId(ws *ws.Listener, id string) {
	s.WsMx.Lock()
	if subs, ok := s.WsMap[ws]; ok {
		delete(s.WsMap[ws], id)
		if len(subs) == 0 {
			delete(s.WsMap, ws)
		}
	}
	s.WsMx.Unlock()
}

// RemoveSubscriber removes a websocket from the subscribers.S collection.
func (s *S) RemoveSubscriber(ws *ws.Listener) {
	s.WsMx.Lock()
	clear(s.WsMap[ws])
	delete(s.WsMap, ws)
	s.WsMx.Unlock()
}

// NotifySubscribers processes a new event and determines whether to send it to subscribers.
func (s *S) NotifySubscribers(authRequired, publicReadable bool, ev *event.T) {
	if ev == nil {
		return
	}
	s.NotifySocketAPI(authRequired, publicReadable, ev)
	s.NotifyHTTPAPI(authRequired, publicReadable, ev)
}
