package json

import (
	"golang.org/x/exp/constraints"

	"realy.mleku.dev/ints"
)

// Unsigned integers have no possible `-` prefix nor a decimal place.
//
// Initialize these values as follows:
//
//	num := &json.Unsigned{}
//
// to get a default value which is zero.
//
// There is a generic NewUnsigned function which will automatically convert any integer type to
// the internal uint64 type, saving the caller from needing to cast it in their code.
type Unsigned struct{ V uint64 }

// NewUnsigned creates a new Unsigned from any unsigned integer.
func NewUnsigned[V constraints.Unsigned](i V) *Unsigned {
	return &Unsigned{uint64(i)}
}

// Marshal an Unsigned into a byte string.
func (u *Unsigned) Marshal(dst []byte) (b []byte) { return ints.New(u.V).Marshal(dst) }

// Unmarshal renders a number in ASCII into an Unsigned.
func (u *Unsigned) Unmarshal(dst []byte) (rem []byte, err error) {
	rem = dst
	n := ints.New(u.V)
	if rem, err = n.Unmarshal(rem); chk.E(err) {
		return
	}
	u.V = n.N
	return
}
