package publisher

import (
	"realy.mleku.dev/event"
	"realy.mleku.dev/typer"
)

type I interface {
	typer.T
	// Deliver the event, accounting for whether auth is required and if the subscriber is
	// authed for protected privacy of privileged messages. if publicReadable, then auth is
	// required if set for writing.
	Deliver(authRequired, publicReadable bool, ev *event.T)
	// Receive accepts a new subscription request, using the typer.T to match it to the
	// publisher.I that handles it.
	Receive(msg typer.T)
}

type Publishers []I
