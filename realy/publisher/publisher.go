// Package publisher is a singleton package that keeps track of subscriptions in
// both websockets and http SSE, including managing the authentication state of
// a connection.
package publisher

import (
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/realy/publisher/openapi"
	"realy.mleku.dev/realy/publisher/socketapi"
)

type API interface {
	Deliver(authRequired, publicReadable bool, ev *event.T)
	ReceiverLoop(ctx context.T)
}

// S is the control structure for the subscription management scheme.
type S struct {
	*socketapi.WSP
	*openapi.HP
}

// New creates a new publisher.S.
func New(ctx context.T) (s *S) {
	s = &S{
		HP:  openapi.NewHP(),
		WSP: socketapi.NewWSP(),
	}
	go s.HP.ReceiverLoop(ctx)
	return
}

func (s *S) Receive(sub any) {
	switch v := sub.(type) {
	case openapi.H:
		s.HP.Chan <- v
	case socketapi.W:
		s.WSP.Chan <- v
	}
}

func (s *S) Deliver(authRequired, publicReadable bool, ev *event.T) {
	s.WSP.Deliver(authRequired, publicReadable, ev)
	s.HP.Deliver(authRequired, publicReadable, ev)
}
