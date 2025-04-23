// Package publisher is a top level router for publishing to registered publishers.
package publish

import (
	"realy.mleku.dev/event"
	"realy.mleku.dev/publish/publisher"
	"realy.mleku.dev/typer"
)

var publishers publisher.Publishers

func Register(p publisher.I) {
	publishers = append(publishers, p)
}

// S is the control structure for the subscription management scheme.
type S struct{ publisher.Publishers }

var _ publisher.I = &S{}

var P = &S{publishers}

func (s *S) Type() string { return "publish" }

func (s *S) Deliver(authRequired, publicReadable bool, ev *event.T) {
	for _, p := range s.Publishers {
		p.Deliver(authRequired, publicReadable, ev)
		return
	}
}

func (s *S) Receive(msg typer.T) {
	t := msg.Type()
	for _, p := range s.Publishers {
		if p.Type() == t {
			p.Receive(msg)
			return
		}
	}
}
