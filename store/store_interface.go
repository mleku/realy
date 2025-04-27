// Package store is an interface and ancillary helpers and types for defining a series of API
// elements for abstracting the event storage from the implementation. It is composed so that
// the top level interface can be partially implemented if need be.
package store

import (
	"io"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/eventidserial"
	"realy.lol/filter"
	"realy.lol/realy/config"
	"realy.lol/tag"
)

// I is an types for a persistence layer for nostr events handled by a relay.
type I interface {
	Initer
	Pather
	io.Closer
	Pather
	Nukener
	Querent
	Deleter
	Saver
	Importer
	Exporter
	Syncer
	LogLeveler
	EventIdSerialer
	Accountant
}

type Initer interface {
	// Init is called at the very beginning by [Server.Start], after [relay.Init], allowing a
	// storage to initialize its internal resources. The parameters can be used by the database
	// implementations to set custom parameters such as cache management and other relevant
	// parameters to the specific implementation.
	Init(path string) (err error)
}

type Pather interface {
	// Path returns the directory of the database.
	Path() (s string)
}

type Nukener interface {
	// Nuke deletes everything in the database.
	Nuke() (err error)
}

type Querent interface {
	// QueryEvents is invoked upon a client's REQ as described in NIP-01. It returns the
	// matching events in reverse chronological order in a slice.
	QueryEvents(c context.T, f *filter.T) (evs event.Ts, err error)
}

type Accountant interface {
	EventCount() (count uint64, err error)
}

type IdTsPk struct {
	Ts  int64
	Id  []byte
	Pub []byte
}

type Querier interface {
	QueryForIds(c context.T, f *filter.T) (evs []IdTsPk, err error)
}

type GetIdsWriter interface {
	FetchIds(w io.Writer, c context.T, evIds *tag.T, binary bool) (err error)
}

type Deleter interface {
	// DeleteEvent is used to handle deletion events, as per NIP-09.
	DeleteEvent(c context.T, ev *eventid.T, noTombstone ...bool) (err error)
}

type Saver interface {
	// SaveEvent is called once relay.AcceptEvent reports true.
	SaveEvent(c context.T, ev *event.T) (err error)
}

type Importer interface {
	// Import reads in a stream of line structured JSON of events to save into the
	// store.
	Import(r io.Reader)
}

type Exporter interface {
	// Export writes a stream of line structured JSON of all events in the store. If pubkeys are
	// present, only those with these pubkeys in the `pubkey` field and in `p` tags will be
	// included.
	Export(c context.T, w io.Writer, pubkeys ...[]byte)
}

type Rescanner interface {
	// Rescan triggers the regeneration of indexes of the database to enable old records to be
	// found with new indexes.
	Rescan() (err error)
}

type Syncer interface {
	// Sync signals the event store to flush its buffers.
	Sync() (err error)
}

type Configuration struct {
	BlockList []string `json:"block_list" doc:"list of IP addresses that will be ignored"`
}

type Configurationer interface {
	GetConfiguration() (c config.C, err error)
	SetConfiguration(c config.C) (err error)
}

type LogLeveler interface {
	SetLogLevel(level string)
}

type EventIdSerialer interface {
	EventIdsBySerial(start uint64, count int) (evs []eventidserial.E,
		err error)
}
