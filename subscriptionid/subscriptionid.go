package subscriptionid

import (
	"crypto/rand"

	"mleku.dev/ec/bech32"
	"mleku.dev/text"
)

type T struct {
	T B
}

func (si *T) String() S { return S(si.T) }

// IsValid returns true if the subscription id is between 1 and 64 characters.
// Invalid means too long or not present.
func (si *T) IsValid() bool { return len(si.T) <= 64 && len(si.T) > 0 }

// New inspects a string and converts to T if it is
// valid. Invalid means length == 0 or length > 64.
func New[V S | B](s V) (*T, error) {
	si := &T{T: B(s)}
	if si.IsValid() {
		return si, nil
	} else {
		// remove invalid return value
		si.T = si.T[:0]
		return si, errorf.E(
			"invalid subscription ID - length %d < 1 or > 64", len(si.T))
	}
}

// MustNew is the same as New except it doesn't check if you feed it rubbish.
func MustNew[V S | B](s V) *T {
	return &T{T: B(s)}
}

const StdLen = 14
const StdHRP = "su"

func NewStd() (t *T) {
	var n int
	var err error
	src := make(B, StdLen)
	if n, err = rand.Read(src); chk.E(err) {
		return
	}
	if n != StdLen {
		err = errorf.E("only read %d of %d bytes from crypto/rand", n, StdLen)
		return
	}
	var bits5 B
	if bits5, err = bech32.ConvertBits(src, 8, 5, true); chk.D(err) {
		return nil
	}
	var dst B
	if dst, err = bech32.Encode(B(StdHRP), bits5); chk.E(err) {
		return
	}
	t = &T{T: dst}
	return
}

func (si *T) MarshalJSON(dst B) (b B, err error) {
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

func (si *T) UnmarshalJSON(b B) (r B, err error) {
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
