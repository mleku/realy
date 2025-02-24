package store

import (
	"realy.lol/envelopes/okenvelope"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/simple"
)

type Methoder interface {
	API() []string
}

type Eventer interface {
	Event(ev *event.T) (err error)
}

type Eventser interface {
	Events(eids []eventid.T) []event.T
}

type Filterer interface {
	Filter(f *simple.Filter) (eids []eventid.T)
}

type Fulltexter interface {
	FullText(ft simple.Fulltext) (eids []eventid.T)
}

type Relayer interface {
	Relay(ev *event.T) (ok *okenvelope.T)
}

type Subscriber interface {
	Subscribe(f *simple.Filter) (eids []eventid.T)
}
type SubscribeFulltexter interface {
	SubscribeFulltext(f *simple.Filter) (eids []eventid.T)
}

type Simple interface {
	Methoder
	Eventer
	Eventser
	Filterer
	Fulltexter
	Relayer
	Subscriber
	SubscribeFulltexter
}
