package prefixes

import (
	"realy.lol/ec/schnorr"
	"realy.lol/errorf"
	"realy.lol/eventid"
	"realy.lol/kind"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/integer"
	"realy.lol/ratel/keys/serial"
	"realy.lol/timestamp"
)

const StartOfWord = index.Len

// these are all offsets from the end
// [ 15 ][ word ][ 32 bytes eventid.T ][ 32 bytes pubkey ][ 8 bytes timestamp.T ][ 2 bytes kind ][ 4 bytes sequence number of word in text ][ 8 bytes Serial ]

const StartOfEventId = eventid.Len + schnorr.PubKeyBytesLen +
	timestamp.Len + kind.Len + integer.Len + serial.Len
const StartOfPubkey = StartOfEventId - eventid.Len
const StartOfTimestamp = StartOfPubkey - schnorr.PubKeyBytesLen
const StartOfKind = StartOfTimestamp - timestamp.Len
const StartOfSequence = StartOfKind - kind.Len
const StartOfSerial = StartOfSequence - integer.Len
const Len = StartOfWord + StartOfEventId

type FulltextIndexKey struct {
	key       []byte
	endOfWord int
	word      []byte
	eventid   *eventid.T
	pubkey    []byte
	timestamp *timestamp.T
	kind      *kind.T
	sequence  uint32
	serial    *serial.T
}

func NewFulltextIndexKey(key []byte) (idx *FulltextIndexKey, err error) {
	if len(key) < Len {
		err = errorf.E("fulltext index key is too short, got %d, minimum is %d", len(key), Len)
		return
	}
	idx = &FulltextIndexKey{key: key, endOfWord: len(key) - StartOfEventId}
	return
}

func (f *FulltextIndexKey) Segment(start, end int) []byte {
	return f.key[len(f.key)-start : len(f.key)-end]
}

func (f *FulltextIndexKey) Word() (v []byte) {
	v = f.key[index.Len:f.endOfWord]
	f.word = v
	return
}

func (f *FulltextIndexKey) EventId() (v *eventid.T) {
	v = eventid.NewWith(f.Segment(StartOfEventId, StartOfPubkey))
	f.eventid = v
	return
}

func (f *FulltextIndexKey) Pubkey() (v []byte) {
	v = f.Segment(StartOfPubkey, StartOfTimestamp)
	f.pubkey = v
	return
}

func (f *FulltextIndexKey) Timestamp() (v *timestamp.T) {
	v = timestamp.FromBytes(f.Segment(StartOfTimestamp, StartOfKind))
	f.timestamp = v
	return
}

func (f *FulltextIndexKey) Kind() (v *kind.T) {
	v = kind.NewFromBytes(f.Segment(StartOfKind, StartOfSequence))
	f.kind = v
	return
}

func (f *FulltextIndexKey) Sequence() (v *integer.T) {
	v = integer.NewFrom(f.Segment(StartOfSequence, StartOfSerial))
	f.sequence = v.Val
	return
}

func (f *FulltextIndexKey) Serial() (v *serial.T) {
	v = serial.New(f.Segment(StartOfSerial, len(f.key)))
	f.serial = v
	return
}
