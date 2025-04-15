// Package publisher is a singleton package that keeps track of subscriptions in
// both websockets and http SSE, including managing the authentication state of
// a connection.
package publish

import (
	"realy.mleku.dev/event"
	"realy.mleku.dev/realy/publish/publisher"
)

// S is the control structure for the subscription management scheme.
type S struct {
	publisher.Publishers
}

// New creates a new publish.S.
func New(p ...publisher.I) (s *S) {
	s = &S{Publishers: p}
	return
}

var _ publisher.I = &S{}

func (s *S) Type() string { return "publish" }

func (s *S) Deliver(authRequired, publicReadable bool, ev *event.T) {
	for _, p := range s.Publishers {
		p.Deliver(authRequired, publicReadable, ev)
		return
	}
}

func (s *S) Receive(msg publisher.Message) {
	t := msg.Type()
	for _, p := range s.Publishers {
		if p.Type() == t {
			p.Receive(msg)
			return
		}
	}
}
