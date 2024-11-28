package json

import (
	"realy.lol/ints"
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

func NewSigned[V int64 | int32 | int16 | int8](i V) *Signed { return &Signed{int64(i)} }

func (s *Signed) Marshal(dst B) (b B) {
	b = dst
	v := s.V
	// we don't add implicit + to the front, only negative
	if v < 0 {
		b = append(b, '-')
		v = -v
	}
	b, _ = ints.New(v).MarshalJSON(b)
	return
}

func (s *Signed) Unmarshal(dst B) (rem B, err E) {
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
	if rem, err = n.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	s.V = int64(n.N)
	// flip sign if neg
	if neg {
		s.V = -s.V
	}
	return
}
