// Package timestamp is a set of helpers for working with timestamps including
// encoding and conversion to various integer forms, from time.Time and varints.
package timestamp

import (
	"encoding/binary"
	"time"
	"unsafe"

	"golang.org/x/exp/constraints"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/ints"
)

// T is a convenience type for UNIX 64 bit timestamps of 1 second
// precision.
type T struct{ V int64 }

// New creates a new timestamp.T, as zero or optionally from teh first variadic parameter as
// int64.
func New[V constraints.Integer](x ...V) (t *T) {
	t = &T{}
	if len(x) > 0 {
		t.V = int64(x[0])
	}
	return
}

// Now returns the current UNIX timestamp of the current second.
func Now() *T {
	tt := T{time.Now().Unix()}
	return &tt
}

// U64 returns the current UNIX timestamp of the current second as uint64.
func (t *T) U64() uint64 {
	// if t == nil {
	// 	return uint64(math.MaxInt64)
	// }
	return uint64(t.V)
}

// I64 returns the current UNIX timestamp of the current second as int64.
func (t *T) I64() int64 {
	// if t == nil {
	// 	return math.MaxInt64
	// }
	return t.V
}

// Time converts a timestamp.Time value into a canonical UNIX 64 bit 1 second
// precision timestamp.
func (t *T) Time() time.Time { return time.Unix(t.V, 0) }

// Int returns the timestamp as an int.
func (t *T) Int() int {
	if t == nil {
		return 0
	}
	return int(t.V)
}

// Bytes returns a timestamp as an 8 byte thing.
func (t *T) Bytes() (b []byte) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t.V))
	return
}

// FromTime returns a T from a time.Time
func FromTime(t time.Time) *T { return &T{t.Unix()} }

// FromUnix converts from a standard int64 unix timestamp.
func FromUnix(t int64) *T { return &T{t} }

func (t *T) FromInt(i int) { *t = T{int64(i)} }

// FromBytes converts from a string of raw bytes.
func FromBytes(b []byte) *T { return &T{int64(binary.BigEndian.Uint64(b))} }

// FromVarint decodes a varint and returns the remainder of the bytes and the encoded
// timestamp.T.
func FromVarint(b []byte) (t *T, rem []byte, err error) {
	n, read := binary.Varint(b)
	if read < 1 {
		err = errorf.E("failed to decode varint timestamp %v", b)
		return
	}
	t = &T{n}
	rem = b[:read]
	return
}

// String renders a timestamp.T as a string.
func (t *T) String() (s string) {
	b := make([]byte, 0, 20)
	tt := ints.New(t.U64())
	b = tt.Marshal(b)
	return unsafe.String(&b[0], len(b))
}

// Marshal a timestamp.T into bytes and append to a provided byte slice.
func (t *T) Marshal(dst []byte) (b []byte) { return ints.New(t.U64()).Marshal(dst) }

// Unmarshal a byte slice with an encoded timestamp.T value and append it to a provided byte
// slice.
func (t *T) Unmarshal(b []byte) (r []byte, err error) {
	n := ints.New(0)
	if r, err = n.Unmarshal(b); chk.E(err) {
		return
	}
	*t = T{n.Int64()}
	return
}

// MarshalJSON marshals a timestamp.T using the json MarshalJSON interface.
func (t *T) MarshalJSON() ([]byte, error) {
	return ints.New(t.U64()).Marshal(nil), nil
}
