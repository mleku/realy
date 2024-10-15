package store

import (
	"io"

	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
)

// I is an types for a persistence layer for nostr events handled by a relay.
type I interface {
	// Init is called at the very beginning by [Server.Start], after [Relay.Init], allowing a
	// storage to initialize its internal resources. The parameters can be used by the database
	// implementations to set custom parameters such as cache management and other relevant
	// parameters to the specific implementation.
	Init(path S) (err E)
	// Path returns the directory of the database.
	Path() S
	// Close must be called after you're done using the store, to free up resources and so on.
	Close() (err E)
	// Nuke deletes everything in the database.
	Nuke() (err E)
	// QueryEvents is invoked upon a client's REQ as described in NIP-01. it should return a
	// channel with the events as they're recovered from a database. the channel should be
	// closed after the events are all delivered.
	QueryEvents(c Ctx, f *filter.T) (evs []*event.T, err E)
	// CountEvents performs the same work as QueryEvents but instead of delivering the events
	// that were found it just returns the count of events
	CountEvents(c Ctx, f *filter.T) (count N, err E)
	// DeleteEvent is used to handle deletion events, as per NIP-09.
	DeleteEvent(c Ctx, ev *eventid.T) (err E)
	// SaveEvent is called once Relay.AcceptEvent reports true.
	SaveEvent(c Ctx, ev *event.T) (err E)
	// Import reads in a stream of line structured JSON of events to save into the store.
	Import(r io.Reader)
	// Export writes a stream of line structured JSON of all events in the store.
	Export(w io.Writer)
}
