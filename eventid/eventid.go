package eventid

import (
	"bytes"
	"errors"

	"lukechampine.com/frand"
	"realy.lol/hex"
	"realy.lol/sha256"
)

// T is the SHA256 hash in hexadecimal of the canonical form of an event as
// produced by the output of T.ToCanonical().Bytes().
type T struct {
	b B
}

func New() (ei *T) { return &T{} }

func NewWith[V S | B](s V) (ei *T) { return &T{b: B(s)} }

func (ei *T) Set(b B) (err E) {
	if len(b) != sha256.Size {
		err = errorf.E("ID bytes incorrect size, got %d require %d", len(b), sha256.Size)
		return
	}
	ei.b = b
	return
}

func NewFromBytes(b B) (ei *T, err E) {
	ei = New()
	if err = ei.Set(b); chk.E(err) {
		return
	}
	return
}

func (ei *T) String() S {
	if ei.b == nil {
		return ""
	}
	return hex.Enc(ei.b)
}

func (ei *T) ByteString(src B) (b B) { return hex.EncAppend(src, ei.b) }

func (ei *T) Bytes() (b B) { return ei.b }

func (ei *T) Len() int {
	if ei == nil {
		log.W.Ln("nil event id")
		return 0
	}
	if ei.b == nil {
		return 0
	}
	return len(ei.b)
}

func (ei *T) Equal(ei2 *T) bool { return bytes.Compare(ei.b, ei2.b) == 0 }

func (ei *T) MarshalJSON() (b B, err E) {
	if ei.b == nil {
		err = errors.New("eventid nil")
		return
	}
	b = make([]byte, 0, 2*sha256.Size+2)
	b = append(b, '"')
	hex.EncAppend(b, ei.b)
	b = append(b, '"')
	return
}

func (ei *T) UnmarshalJSON(b B) (err E) {
	if len(ei.b) != sha256.Size {
		ei.b = make([]byte, 0, sha256.Size)
	}
	b = b[1 : 2*sha256.Size+1]
	if len(b) != 2*sha256.Size {
		err = errorf.E("event ID hex incorrect size, got %d require %d",
			len(b), 2*sha256.Size)
		log.E.Ln(string(b))
		return
	}
	ei.b = make([]byte, 0, sha256.Size)
	ei.b, err = hex.DecAppend(ei.b, b)
	return
}

// NewFromString inspects a string and ensures it is a valid, 64 character long
// hexadecimal string, returns the string coerced to the type.
func NewFromString(s S) (ei *T, err E) {
	if len(s) != 2*sha256.Size {
		return nil, errorf.E("event ID hex wrong size, got %d require %d",
			len(s), 2*sha256.Size)
	}
	ei = &T{b: make([]byte, 0, sha256.Size)}
	ei.b, err = hex.DecAppend(ei.b, []byte(s))
	return
}

// Gen creates a fake pseudorandom generated event ID for tests.
func Gen() (ei *T) { return &T{frand.Bytes(sha256.Size)} }
