// Package eventid is a codec for managing nostr event Ids (hash of the
// canonical form of a nostr event).
package eventid

import (
	"lukechampine.com/frand"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/hex"
	"realy.lol/log"
	"realy.lol/sha256"
)

// T is the SHA256 hash in hexadecimal of the canonical form of an event as
// produced by the output of T.ToCanonical().Bytes().
type T [sha256.Size]byte

// New creates a new eventid.T. This is actually more wordy than simply creating a &T{} via
// slice literal.
func New() (ei *T) { return &T{} }

// NewWith creates an eventid.T out of bytes or string but assumes it is binary
// and that it is the right length. The result is either truncated or padded automatically by
// the use of the "copy" operation.
func NewWith[V string | []byte](s V) (ei *T) {
	id := T{}
	copy(id[:], s)
	return &id
}

// Set the value of an eventid.T with checking of the length before copying it.
func (ei *T) Set(b []byte) (err error) {
	if ei == nil {
		err = errorf.E("event id is nil")
		return
	}
	if len(b) != sha256.Size {
		err = errorf.E("Id bytes incorrect size, got %d require %d",
			len(b), sha256.Size)
		return
	}
	copy(ei[:], b)
	return
}

// NewFromBytes creates a new eventid.T from the raw event Id hash.
func NewFromBytes(b []byte) (ei *T, err error) {
	ei = New()
	if err = ei.Set(b); chk.E(err) {
		return
	}
	return
}

// String renders an eventid.T as a string.
func (ei *T) String() string {
	if ei == nil {
		return ""
	}
	return hex.Enc(ei[:])
}

// ByteString renders an eventid.T as bytes in ASCII hex.
func (ei *T) ByteString(src []byte) (b []byte) {
	return hex.EncAppend(src, ei[:])
}

// Bytes returns the raw bytes of the eventid.T.
func (ei *T) Bytes() (b []byte) { return ei[:] }

// Len returns the length of the eventid.T.
func (ei *T) Len() int {
	if ei == nil {
		log.W.Ln("nil event id")
		return 0
	}
	return len(ei)
}

// Equal tests whether another eventid.T is the same.
func (ei *T) Equal(ei2 *T) (eq bool) {
	if ei == nil || ei2 == nil {
		log.W.Ln("can't compare to nil event id")
		return
	}
	return *ei == *ei2
}

// Marshal renders the eventid.T into JSON.
func (ei *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = make([]byte, 0, 2*sha256.Size+2)
	b = append(b, '"')
	hex.EncAppend(b, ei[:])
	b = append(b, '"')
	return
}

// Unmarshal decodes a JSON encoded eventid.T.
func (ei *T) Unmarshal(b []byte) (rem []byte, err error) {
	// trim off the quotes.
	b = b[1 : 2*sha256.Size+1]
	if len(b) != 2*sha256.Size {
		err = errorf.E("event Id hex incorrect size, got %d require %d",
			len(b), 2*sha256.Size)
		log.E.Ln(string(b))
		return
	}
	var bb []byte
	if bb, err = hex.Dec(string(b)); chk.E(err) {
		return
	}
	copy(ei[:], bb)
	return
}

// NewFromString inspects a string and ensures it is a valid, 64 character long
// hexadecimal string, returns the string coerced to the type.
func NewFromString(s string) (ei *T, err error) {
	if len(s) != 2*sha256.Size {
		return nil, errorf.E("event Id hex wrong size, got %d require %d",
			len(s), 2*sha256.Size)
	}
	ei = &T{}
	b := make([]byte, 0, sha256.Size)
	b, err = hex.DecAppend(b, []byte(s))
	copy(ei[:], b)
	return
}

// Gen creates a fake pseudorandom generated event Id for tests.
func Gen() (ei *T) {
	b := frand.Bytes(sha256.Size)
	ei = &T{}
	copy(ei[:], b)
	return
}
