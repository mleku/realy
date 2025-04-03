// Package hex is a set of aliases and helpers for using the templexxx SIMD hex
// encoder.
package hex

import (
	"encoding/hex"

	"github.com/templexxx/xhex"
)

var Enc = hex.EncodeToString
var EncBytes = hex.Encode
var Dec = hex.DecodeString
var DecBytes = hex.Decode

// var EncAppend = hex.AppendEncode
// var DecAppend = hex.AppendDecode

var DecLen = hex.DecodedLen

type InvalidByteError = hex.InvalidByteError

// EncAppend uses xhex to encode a sice of bytes and appends it to a provided destination slice.
func EncAppend(dst, src []byte) (b []byte) {
	l := len(dst)
	dst = append(dst, make([]byte, len(src)*2)...)
	xhex.Encode(dst[l:], src)
	return dst
}

// DecAppend decodes a provided encoded hex encoded string and appends the decoded output to a
// provided input slice.
func DecAppend(dst, src []byte) (b []byte, err error) {
	if src == nil || len(src) < 2 {
		err = errorf.E("nothing to decode")
		return
	}
	if dst == nil {
		dst = []byte{}
	}
	l := len(dst)
	b = dst
	b = append(b, make([]byte, len(src)/2)...)
	if err = xhex.Decode(b[l:], src); chk.T(err) {
		return
	}
	return
}
