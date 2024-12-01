package timestamp

import (
	"encoding/binary"
	"time"
	"unsafe"

	"realy.lol/ints"
)

// T is a convenience type for UNIX 64 bit timestamps of 1 second
// precision.
type T int64

func New() (t *T) {
	tt := T(0)
	return &tt
}

// Now returns the current UNIX timestamp of the current second.
func Now() *T {
	tt := T(time.Now().Unix())
	return &tt
}

// U64 returns the current UNIX timestamp of the current second as uint64.
func (t *T) U64() uint64 {
	if t == nil {
		return 0
	}
	return uint64(*t)
}

// I64 returns the current UNIX timestamp of the current second as int64.
func (t *T) I64() int64 {
	if t == nil {
		return 0
	}
	return int64(*t)
}

// Time converts a timestamp.Time value into a canonical UNIX 64 bit 1 second
// precision timestamp.
func (t *T) Time() time.Time { return time.Unix(int64(*t), 0) }

// Int returns the timestamp as an int.
func (t *T) Int() no {
	if t == nil {
		return 0
	}
	return no(*t)
}

func (t *T) Bytes() (b by) {
	b = make(by, 8)
	binary.BigEndian.PutUint64(b, uint64(*t))
	return
}

// FromTime returns a T from a time.Time
func FromTime(t time.Time) *T {
	tt := T(t.Unix())
	return &tt
}

// FromUnix converts from a standard int64 unix timestamp.
func FromUnix(t int64) *T {
	tt := T(t)
	return &tt
}
func (t *T) FromInt(i no) { *t = T(i) }

// FromBytes converts from a string of raw bytes.
func FromBytes(b by) *T {
	tt := T(binary.BigEndian.Uint64(b))
	return &tt
}

func FromVarint(b by) (t *T, rem by, err er) {
	n, read := binary.Varint(b)
	if read < 1 {
		err = errorf.E("failed to decode varint timestamp %v", b)
		return
	}
	tt := T(n)
	t = &tt
	rem = b[:read]
	return
}

func ToVarint(dst by, t *T) by { return binary.AppendVarint(dst, int64(*t)) }

func (t *T) FromVarint(dst by) (b by) { return ToVarint(dst, t) }

func (t *T) String() (s st) {
	b := make(by, 0, 20)
	var err er
	tt := ints.New(t.U64())
	if b, err = tt.MarshalJSON(b); chk.E(err) {
		return
	}
	return unsafe.String(&b[0], len(b))
}

func (t *T) MarshalJSON(dst by) (b by, err er) {
	tt := ints.New(t.U64())
	return tt.MarshalJSON(dst)
}

func (t *T) UnmarshalJSON(b by) (r by, err er) {
	n := ints.New(0)
	r, err = n.UnmarshalJSON(b)
	*t = T(n.Uint64())
	return
}
