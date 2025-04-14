package openapi

import (
	"bytes"
	"sync"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/tag"
)

// H is the control structure for a HTTP SSE subscription, including the filter, authed
// pubkey and a channel to send the events to.
type H struct {
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

// Map is a collection of H TTP subscriptions.
type Map map[*H]struct{}

type HP struct {
	// Map is the map of subscriptions from the http api.
	Map
	// Chan is a channel that http api subscriptions send their receiver channel through.
	Chan chan H
	// HLock is the mutex that locks the Map.
	Mx sync.Mutex
}

func NewHP() *HP {
	return &HP{
		Map:  make(Map),
		Chan: make(chan H),
	}
}

func (hp *HP) ReceiverLoop(ctx context.T) {
	for {
		select {
		case <-ctx.Done():
			return
		case h := <-hp.Chan:
			hp.Mx.Lock()
			hp.Map[&h] = struct{}{}
			hp.Mx.Unlock()
		}
	}
}

func (hp *HP) Deliver(authRequired, publicReadable bool, ev *event.T) {
	hp.Mx.Lock()
	var subs []*H
	for sub := range hp.Map {
		// check if the subscription'hp subscriber is still alive
		select {
		case <-sub.Ctx.Done():
			subs = append(subs, sub)
		default:
		}
	}
	for _, sub := range subs {
		delete(hp.Map, sub)
	}
	subs = subs[:0]
	for sub := range hp.Map {
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
	hp.Mx.Unlock()
}
