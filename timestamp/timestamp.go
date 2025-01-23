package timestamp

import (
	"encoding/binary"
	"time"
	"unsafe"

	"realy.lol/ints"
)

// T is a convenience type for UNIX 64 bit timestamps of 1 second
// precision.
type T struct{ V int64 }

func New() (t *T) { return &T{} }

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
func (t *T) Int() no {
	if t == nil {
		return 0
	}
	return no(t.V)
}

func (t *T) Bytes() (b by) {
	b = make(by, 8)
	binary.BigEndian.PutUint64(b, uint64(t.V))
	return
}

// FromTime returns a T from a time.Time
func FromTime(t time.Time) *T { return &T{t.Unix()} }

// FromUnix converts from a standard int64 unix timestamp.
func FromUnix(t int64) *T { return &T{t} }

func (t *T) FromInt(i no) { *t = T{int64(i)} }

// FromBytes converts from a string of raw bytes.
func FromBytes(b by) *T { return &T{int64(binary.BigEndian.Uint64(b))} }

func FromVarint(b by) (t *T, rem by, err er) {
	n, read := binary.Varint(b)
	if read < 1 {
		err = errorf.E("failed to decode varint timestamp %v", b)
		return
	}
	t = &T{n}
	rem = b[:read]
	return
}

func ToVarint(dst by, t *T) by { return binary.AppendVarint(dst, t.V) }

func (t *T) FromVarint(dst by) (b by) { return ToVarint(dst, t) }

func (t *T) String() (s st) {
	b := make(by, 0, 20)
	tt := ints.New(t.U64())
	b = tt.Marshal(b)
	return unsafe.String(&b[0], len(b))
}

func (t *T) Marshal(dst by) (b by) { return ints.New(t.U64()).Marshal(dst) }

func (t *T) Unmarshal(b by) (r by, err er) {
	n := ints.New(0)
	r, err = n.Unmarshal(b)
	*t = T{n.Int64()}
	return
}

func (t *T) MarshalJSON() ([]byte, error) {
	return ints.New(t.U64()).Marshal(nil), nil
}
