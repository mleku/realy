package json

import (
	"golang.org/x/exp/constraints"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/ints"
)

// Signed integers can be negative and thus a `-` prefix.
//
// Initialize these values as follows:
//
//	num := &json.Signed{}
//
// to get a default value which is zero.
//
// Technically we are supporting negative numbers here but in fact in nostr message encodings
// there is no such thing as any negative integers, nor floating point, but this ensures
// if there is ever need for negative values they are supported. For the most part this will be
// used for timestamps.
//
// There is a NewSigned function that accepts any type of generic signed integer value and
// automatically converts it to the biggest type that is used in runtime.
type Signed struct{ V int64 }

// NewSigned creates a new Signed integer value.
func NewSigned[V constraints.Signed](i V) *Signed { return &Signed{int64(i)} }

// Marshal the Signed into a byte string in standard JSON formatting.
func (s *Signed) Marshal(dst []byte) (b []byte) {
	b = dst
	v := s.V
	// we don't add implicit + to the front, only negative
	if v < 0 {
		b = append(b, '-')
		v = -v
	}
	b = ints.New(v).Marshal(b)
	return
}

// Unmarshal a Signed in JSON form into its value.
func (s *Signed) Unmarshal(dst []byte) (rem []byte, err error) {
	rem = dst
	var neg bool
	// first byte can be `-` or `+`
	if rem[0] == '-' {
		neg = true
		// advance to next
		rem = rem[1:]
	} else if rem[0] == '+' {
		rem = rem[1:]
	}
	n := &ints.T{}
	if rem, err = n.Unmarshal(rem); chk.E(err) {
		return
	}
	s.V = int64(n.N)
	// flip sign if neg
	if neg {
		s.V = -s.V
	}
	return
}
