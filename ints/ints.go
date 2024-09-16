package ints

import (
	_ "embed"
)

// run this to regenerate (pointlessly) the base 10 array of 4 places per entry
//go:generate go run ./gen/.

//go:embed base10k.txt
var base10k []byte

const base = 10000

type T struct {
	N uint64
}

func New[V int | uint64 | uint32 | uint16 | uint8 | int64 | int32 | int16 | int8](n V) *T {
	return &T{uint64(n)}
}

func (n *T) Uint64() uint64 { return n.N }
func (n *T) Uint16() uint16 { return uint16(n.N) }

var powers = []*T{
	{1},
	{1_0000},
	{1_0000_0000},
	{1_0000_0000_0000},
	{1_0000_0000_0000_0000},
}

const zero = '0'
const nine = '9'

// MarshalJSON encodes an uint64 into ASCII decimal format in a
// []byte.
func (n *T) MarshalJSON(dst B) (b B, err error) {
	nn := n.N
	b = dst
	if n.N == 0 {
		b = append(b, '0')
		return
	}
	var i int
	var trimmed bool
	k := len(powers)
	for k > 0 {
		k--
		q := n.N / powers[k].N
		if !trimmed && q == 0 {
			continue
		}
		offset := q * 4
		bb := base10k[offset : offset+4]
		if !trimmed {
			for i = range bb {
				if bb[i] != '0' {
					bb = bb[i:]
					trimmed = true
					break
				}
			}
		}
		b = append(b, bb...)
		n.N = n.N - q*powers[k].N
	}
	n.N = nn
	return
}

// UnmarshalJSON reads a string, which must be a positive integer no larger than math.MaxUint64,
// skipping any non-numeric content before it.
//
// Note that leading zeros are not considered valid, but basically no such thing as machine
// generated JSON integers with leading zeroes. Until this is disproven, this is the fastest way
// to read a positive json integer, and a leading zero is decoded as a zero, and the remainder
// returned.
func (n *T) UnmarshalJSON(b B) (r B, err error) {
	var sLen int
	if b[0] == zero {
		r = b[1:]
		n.N = 0
		return
	}
	// count the digits
	for ; sLen < len(b) && b[sLen] >= zero && b[sLen] <= nine && b[sLen] != ','; sLen++ {
	}
	if sLen == 0 {
		err = errorf.E("zero length number")
		return
	}
	if sLen > 20 {
		err = errorf.E("too big number for uint64")
		return
	}
	// the length of the string found
	r = b[sLen:]
	b = b[:sLen]
	for _, ch := range b {
		ch -= zero
		n.N = n.N*10 + uint64(ch)
	}
	return
}
