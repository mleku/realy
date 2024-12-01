package json

import (
	"realy.lol/ints"
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

func NewUnsigned[V int64 | int32 | int16 | int8 | uint64 | uint32 | uint16 |
	uint8](i V) *Signed {

	return &Signed{int64(i)}
}

func (u *Unsigned) Marshal(dst by) (b by) {
	b, _ = ints.New(u.V).MarshalJSON(dst)
	return
}

func (u *Unsigned) Unmarshal(dst by) (rem by, err er) {
	rem = dst
	n := ints.New(u.V)
	if rem, err = n.UnmarshalJSON(dst); chk.E(err) {
		return
	}
	u.V = n.N
	return
}
