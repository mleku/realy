package subscription

import (
	"crypto/rand"

	"realy.lol/ec/bech32"
	"realy.lol/text"
)

type Id struct {
	T by
}

func (si *Id) String() st { return st(si.T) }

// IsValid returns true if the subscription id is between 1 and 64 characters.
// Invalid means too long or not present.
func (si *Id) IsValid() bo { return len(si.T) <= 64 && len(si.T) > 0 }

// NewId inspects a string and converts to Id if it is
// valid. Invalid means length == 0 or length > 64.
func NewId[V st | by](s V) (*Id, er) {
	si := &Id{T: by(s)}
	if si.IsValid() {
		return si, nil
	} else {
		// remove invalid return value
		si.T = si.T[:0]
		return si, errorf.E(
			"invalid subscription ID - length %d < 1 or > 64", len(si.T))
	}
}

// MustNew is the same as NewId except it doesn't check if you feed it rubbish.
//
// DO NOT USE WITHOUT CHECKING THE ID IS NOT NIL AND > 0 AND <= 64
func MustNew[V st | by](s V) *Id {
	return &Id{T: by(s)}
}

const StdLen = 14
const StdHRP = "su"

func NewStd() (t *Id) {
	var n no
	var err er
	src := make(by, StdLen)
	if n, err = rand.Read(src); chk.E(err) {
		return
	}
	if n != StdLen {
		err = errorf.E("only read %d of %d bytes from crypto/rand", n, StdLen)
		return
	}
	var bits5 by
	if bits5, err = bech32.ConvertBits(src, 8, 5, true); chk.D(err) {
		return nil
	}
	var dst by
	if dst, err = bech32.Encode(by(StdHRP), bits5); chk.E(err) {
		return
	}
	t = &Id{T: dst}
	return
}

func (si *Id) MarshalJSON(dst by) (b by, err er) {
	ue := text.NostrEscape(nil, si.T)
	if len(ue) < 1 || len(ue) > 64 {
		err = errorf.E("invalid subscription ID, must be between 1 and 64 "+
			"characters, got %d (possibly due to escaping)", len(ue))
		return
	}
	b = dst
	b = append(b, '"')
	b = append(b, ue...)
	b = append(b, '"')
	return
}

func (si *Id) UnmarshalJSON(b by) (r by, err er) {
	var openQuotes, escaping bo
	var start no
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
