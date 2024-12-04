package index

import (
	"realy.lol/ratel/keys"
)

type P byte

// Key writes a key with the P prefix byte and an arbitrary list of
// keys.Element.
func (p P) Key(element ...keys.Element) (b by) {
	b = keys.Write(
		append([]keys.Element{New(byte(p))}, element...)...)
	// log.T.F("key %x", b)
	return
}

// B returns the index.P as a byte.
func (p P) B() byte { return byte(p) }

// I returns the index.P as an int (for use with the KeySizes.
func (p P) I() no { return no(p) }

// GetAsBytes todo wat is dis?
func GetAsBytes(prf ...P) (b []by) {
	b = make([]by, len(prf))
	for i := range prf {
		b[i] = by{byte(prf[i])}
	}
	return
}

const (
	// Version is the key that stores the version number, the value is a 16-bit
	// integer (2 bytes)
	//
	//   [ 255 ][ 2 byte/16 bit version code ]
	Version P = 255
)
const (
	// Event is the prefix used with a Serial counter value provided by badgerDB to
	// provide conflict-free 8 byte 64-bit unique keys for event records, which
	// follows the prefix.
	//
	//   [ 0 ][ 8 bytes Serial ]
	Event P = iota

	// CreatedAt creates an index key that contains the unix
	// timestamp of the event record serial.
	//
	//   [ 1 ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	CreatedAt

	// Id contains the first 8 bytes of the ID of the event and the 8
	// byte Serial of the event record.
	//
	//   [ 2 ][ 8 bytes eventid.T prefix ][ 8 bytes Serial ]
	Id

	// Kind contains the kind and datestamp.
	//
	//   [ 3 ][ 2 bytes kind.T ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Kind

	// Pubkey contains pubkey prefix and timestamp.
	//
	//   [ 4 ][ 8 bytes pubkey prefix ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Pubkey

	// PubkeyKind contains pubkey prefix, kind and timestamp.
	//
	//   [ 5 ][ 8 bytes pubkey prefix ][ 2 bytes kind.T ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	PubkeyKind

	// Tag is for miscellaneous arbitrary length tags, with timestamp and event
	// serial after.
	//
	//   [ 6 ][ tag string 1 <= 100 bytes ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Tag

	// Tag32 contains the 8 byte pubkey prefix, timestamp and serial.
	//
	//   [ 7 ][ 8 bytes pubkey prefix ][ 8 bytes timestamp.T ][ 8 bytes Serial ]
	Tag32

	// TagAddr contains the kind, pubkey prefix, value (index 2) of address tag (eg
	// relay address), followed by timestamp and serial.
	//
	//   [ 8 ][ 2 byte kind.T][ 8 byte pubkey prefix ][ network address ][ 8 byte timestamp.T ][ 8 byte Serial ]
	TagAddr

	// Counter is the eventid.T prefix, value stores the average time of access
	// (average of all access timestamps) and the size of the record.
	//
	//   [ 9 ][ 8 bytes Serial ] : value: [ 8 bytes timestamp ]
	Counter

	// Tombstone is an index that contains the left half of an event ID that has
	// been deleted. The purpose of this event is to stop the event being
	// republished, as a delete event may not be respected by other relays and
	// eventually lead to a republication. The timestamp is added at the end to
	// enable pruning the oldest tombstones.
	//
	// [ 10 ][ 16 bytes first/left half of event ID ][ 8 bytes timestamp ]
	Tombstone
)

// FilterPrefixes is a slice of the prefixes used by filter index to enable a loop
// for pulling events matching a serial
var FilterPrefixes = []by{
	{CreatedAt.B()},
	{Id.B()},
	{Kind.B()},
	{Pubkey.B()},
	{PubkeyKind.B()},
	{Tag.B()},
	{Tag32.B()},
	{TagAddr.B()},
}
