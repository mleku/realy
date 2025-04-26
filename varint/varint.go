// Package varint is a variable integer encoding that works in reverse compared to the stdlib
// binary Varint. The terminal byte in the encoding is the one with the 8th bit set. This is
// basically like a base 128 encoding. It reads forward using an io.Reader and writes forward
// using an io.Writer.
package varint

import (
	"io"

	"realy.lol/chk"
)

func Encode(w io.Writer, v uint64) {
	x := []byte{0}
	for {
		x[0] = byte(v) & 127
		v >>= 7
		if v == 0 {
			x[0] |= 128
			_, _ = w.Write(x)
			break
		} else {
			_, _ = w.Write(x)
		}
	}
}

func Decode(r io.Reader) (v uint64, err error) {
	x := []byte{0}
	// _, _ = r.Read(x)
	// if x[0] > 128 {
	// 	v += uint64(x[0] & 127)
	// 	return
	// } else {
	v += uint64(x[0])
	var i uint64
	for {
		if _, err = r.Read(x); chk.E(err) {
			return
		}
		if x[0] >= 128 {
			v += uint64(x[0]&127) << (i * 7)
			return
		} else {
			v += uint64(x[0]) << (i * 7)
			i++
		}
	}
	// }
}
