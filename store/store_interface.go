package store

import (
	"io"

	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
)

// I is an types for a persistence layer for nostr events handled by a relay.
type I interface {
	Initializer
	Pather
	// Closer must be called after you're done using the store, to free up resources
	// and so on.
	io.Closer
	Nukener
	Querent
	Counter
	Deleter
	Saver
	Importer
	Exporter
	Syncer
}

type Initializer interface {
	// Init is called at the very beginning by [Server.Start], after [Relay.Init],
	// allowing a storage to initialize its internal resources. The parameters can be
	// used by the database implementations to set custom parameters such as cache
	// management and other relevant parameters to the specific implementation.
	Init(path st) (err er)
}

type Pather interface {
	// Path returns the directory of the database.
	Path() (s st)
}

type Nukener interface {
	// Nuke deletes everything in the database.
	Nuke() (err er)
}

type Querent interface {
	// QueryEvents is invoked upon a client's REQ as described in NIP-01. It returns
	// the matching events in reverse chronological order in a slice.
	//
	// if ours is set, this means that limits applying to external clients are
	// not in force (eg maxlimit).
	QueryEvents(c cx, f *filter.T, ours ...bo) (evs event.Ts, err er)
}

type Counter interface {
	// CountEvents performs the same work as QueryEvents but instead of delivering
	// the events that were found it just returns the count of events
	CountEvents(c cx, f *filter.T) (count no, approx bo, err er)
}

type Deleter interface {
	// DeleteEvent is used to handle deletion events, as per NIP-09.
	DeleteEvent(c cx, ev *eventid.T) (err er)
}

type Saver interface {
	// SaveEvent is called once Relay.AcceptEvent reports true.
	SaveEvent(c cx, ev *event.T) (err er)
}

type Importer interface {
	// Import reads in a stream of line structured JSON of events to save into the
	// store.
	Import(r io.Reader)
}

type Exporter interface {
	// Export writes a stream of line structured JSON of all events in the store. If
	// pubkeys are present, only those with these pubkeys in the `pubkey` field and
	// in `p` tags will be included.
	Export(c cx, w io.Writer, pubkeys ...by)
}

type Syncer interface {
	// Sync signals the event store to flush its buffers.
	Sync() (err er)
}
