package publisher

import (
	"realy.mleku.dev/event"
)

type Message interface {
	Type() string
}

type I interface {
	Message
	Deliver(authRequired, publicReadable bool, ev *event.T)
	Receive(msg Message)
}

type Publishers []I
