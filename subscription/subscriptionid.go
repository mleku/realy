// Package subscription is a set of helpers for managing nostr websocket
// subscription Ids, used with the REQ method to maintain an association between
// a REQ and resultant messages such as EVENT and CLOSED.
package subscription

import (
	"crypto/rand"

	"realy.mleku.dev/ec/bech32"
	"realy.mleku.dev/text"
)

type Id struct {
	T []byte
}

func (si *Id) String() string { return string(si.T) }

// IsValid returns true if the subscription id is between 1 and 64 characters.
// Invalid means too long or not present.
func (si *Id) IsValid() bool { return len(si.T) <= 64 && len(si.T) > 0 }

// NewId inspects a string and converts to Id if it is
// valid. Invalid means length == 0 or length > 64.
func NewId[V string | []byte](s V) (*Id, error) {
	si := &Id{T: []byte(s)}
	if si.IsValid() {
		return si, nil
	} else {
		// remove invalid return value
		si.T = si.T[:0]
		return si, errorf.E(
			"invalid subscription Id - length %d < 1 or > 64", len(si.T))
	}
}

// MustNew is the same as NewId except it doesn't check if you feed it rubbish.
//
// DO NOT USE WITHOUT CHECKING THE Id IS NOT NIL AND > 0 AND <= 64
func MustNew[V string | []byte](s V) *Id {
	return &Id{T: []byte(s)}
}

const StdLen = 14
const StdHRP = "su"

// NewStd creates a new standard subscription ID, which is a 14 byte long (92 bit) identifier,
// encoded using bech32.
func NewStd() (t *Id) {
	var n int
	var err error
	src := make([]byte, StdLen)
	if n, err = rand.Read(src); chk.E(err) {
		return
	}
	if n != StdLen {
		err = errorf.E("only read %d of %d bytes from crypto/rand", n, StdLen)
		return
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(src, 8, 5, true); chk.D(err) {
		return nil
	}
	var dst []byte
	if dst, err = bech32.Encode([]byte(StdHRP), bits5); chk.E(err) {
		return
	}
	t = &Id{T: dst}
	return
}

// Marshal renders the raw bytes of a subscription.Id to raw byte string.
func (si *Id) Marshal(dst []byte) (b []byte) {
	ue := text.NostrEscape(nil, si.T)
	if len(ue) < 1 || len(ue) > 64 {
		log.E.F("invalid subscription Id, must be between 1 and 64 "+
			"characters, got %d (possibly due to escaping)", len(ue))
		return
	}
	b = dst
	b = append(b, '"')
	b = append(b, ue...)
	b = append(b, '"')
	return
}

// Unmarshal a subscription.Id from raw bytes.
func (si *Id) Unmarshal(b []byte) (r []byte, err error) {
	var openQuotes, escaping bool
	var start int
	r = b
	for i := range r {
		if !openQuotes && r[i] == '"' {
			openQuotes = true
			start = i + 1
		} else if openQuotes {
			if !escaping && r[i] == '\\' {
				escaping = true
			} else if r[i] == '"' {
				if !escaping {
					si.T = text.NostrUnescape(r[start:i])
					r = r[i+1:]
					return
				} else {
					escaping = false
				}
			} else {
				escaping = false
			}
		}
	}
	return
}
