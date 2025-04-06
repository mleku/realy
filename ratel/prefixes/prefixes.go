// Package prefixes provides a list of the index.P types that designate tables
// in the ratel event store, as well as enabling a simple syntax to assemble and
// decompose an index key into its keys.Element s.
package prefixes

import (
	"realy.lol/ec/schnorr"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
)

const (
	// Version is the key that stores the version number, the value is a 16-bit
	// integer (2 bytes)
	//
	//   [ 255 ][ 2 byte/16 bit version code ]
	Version index.P = 255

	// Event is the prefix used with a Serial counter value provided by badgerDB to
	// provide conflict-free 8 byte 64-bit unique keys for event records, which
	// follows the prefix.
	//
	//   [ 1 ][ 8 bytes Serial ]
	Event index.P = iota

	// CreatedAt creates an index key that contains the unix
	// timestamp of the event record serial.
	//
	//   [ 2 ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	CreatedAt

	// Id contains the first 8 bytes of the Id of the event and the 8
	// byte Serial of the event record.
	//
	//   [ 3 ][ 8 bytes eventid.T prefix ][ 8 bytes Serial ]
	Id

	// Kind contains the kind and datestamp.
	//
	//   [ 4 ][ 2 bytes kind.T ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Kind

	// Pubkey contains pubkey prefix and timestamp.
	//
	//   [ 5 ][ 8 bytes pubkey prefix ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Pubkey

	// PubkeyKind contains pubkey prefix, kind and timestamp.
	//
	//   [ 6 ][ 8 bytes pubkey prefix ][ 2 bytes kind.T ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	PubkeyKind

	// Tag is for miscellaneous arbitrary length tags, with timestamp and event
	// serial after.
	//
	//   [ 7 ][ tag string 1 <= 100 bytes ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Tag

	// Tag32 contains the 8 byte pubkey prefix, timestamp and serial.
	//
	//   [ 8 ][ 8 bytes pubkey prefix ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Tag32

	// TagAddr contains the kind, pubkey prefix, value (index 2) of address tag (eg
	// relay address), followed by timestamp and serial.
	//
	//   [ 9 ][ 2 byte kind.T][ 8 byte pubkey prefix ][ network address ][ 8 byte timestamp.T ][ 8 byte Serial ]
	TagAddr

	// Counter is the eventid.T prefix, value stores the average time of access
	// (average of all access timestamps) and the size of the record.
	//
	//   [ 10 ][ 8 bytes Serial ] : value: [ 8 bytes timestamp ]
	Counter

	// Tombstone is an index that contains the left half of an event Id that has
	// been deleted. The purpose of this event is to stop the event being
	// republished, as a delete event may not be respected by other relays and
	// eventually lead to a republication. The timestamp is added at the end to
	// enable pruning the oldest tombstones.
	//
	// [ 11 ][ 16 bytes first/left half of event Id ][ 8 bytes timestamp ]
	Tombstone

	// PubkeyIndex is the prefix for an index that stores a mapping between pubkeys
	// and a pubkey serial.
	//
	// todo: this is useful feature but rather than for saving space on pubkeys in
	//       events might have a more useful place in some kind of search API. eg just
	//       want pubkey from event id, combined with FullIndex.
	//
	// [ 12 ][ 32 bytes pubkey ][ 8 bytes pubkey serial ]
	PubkeyIndex

	// FullIndex is a secondary table for IDs that is used to fetch the full Id
	// hash instead of fetching and unmarshalling the event. The Id index will
	// ultimately be deprecated in favor of this because returning event Ids and
	// letting the client handle pagination reduces relay complexity.
	//
	// In addition, as a mechanism of sorting, the event Id bears also a timestamp
	// from its created_at field. The serial acts as a "first seen" ordering, then
	// you also have the (claimed) chronological ordering.
	//
	//   [ 13 ][ 8 bytes Serial ][ 32 bytes eventid.T ][ 32 bytes pubkey ][ 8 bytes timestamp.T ]
	FullIndex

	// Configuration is a free-form minified JSON object that contains a collection of
	// configuration items.
	//
	// [ 14 ]
	Configuration
)

// FilterPrefixes is a slice of the prefixes used by filter index to enable a loop
// for pulling events matching a serial
var FilterPrefixes = [][]byte{
	{CreatedAt.B()},
	{Id.B()},
	{Kind.B()},
	{Pubkey.B()},
	{PubkeyKind.B()},
	{Tag.B()},
	{Tag32.B()},
	{TagAddr.B()},
	{FullIndex.B()},
}

// AllPrefixes is used to do a full database nuke.
var AllPrefixes = [][]byte{
	{Event.B()},
	{CreatedAt.B()},
	{Id.B()},
	{Kind.B()},
	{Pubkey.B()},
	{PubkeyKind.B()},
	{Tag.B()},
	{Tag32.B()},
	{TagAddr.B()},
	{Counter.B()},
	{PubkeyIndex.B()},
	{FullIndex.B()},
}

// KeySizes are the byte size of keys of each type of key prefix. int(P) or call the P.I() method
// corresponds to the index 1:1. For future index additions be sure to add the
// relevant KeySizes sum as it describes the data for a programmer.
var KeySizes = []int{
	// Event
	1 + serial.Len,
	// CreatedAt
	1 + createdat.Len + serial.Len,
	// Id
	1 + id.Len + serial.Len,
	// Kind
	1 + kinder.Len + createdat.Len + serial.Len,
	// Pubkey
	1 + pubkey.Len + createdat.Len + serial.Len,
	// PubkeyKind
	1 + pubkey.Len + kinder.Len + createdat.Len + serial.Len,
	// Tag (worst case scenario)
	1 + 100 + createdat.Len + serial.Len,
	// Tag32
	1 + pubkey.Len + createdat.Len + serial.Len,
	// TagAddr
	1 + kinder.Len + pubkey.Len + 100 + createdat.Len + serial.Len,
	// Counter
	1 + serial.Len,
	// Tombstone
	1 + sha256.Size/2 + serial.Len,
	// PubkeyIndex
	1 + schnorr.PubKeyBytesLen + serial.Len,
	// FullIndex
	1 + fullid.Len + createdat.Len + serial.Len,
}
