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

func EncAppend(dst, src B) (b B) {
	l := len(dst)
	dst = append(dst, make(B, len(src)*2)...)
	xhex.Encode(dst[l:], src)
	return dst
}

func DecAppend(dst, src B) (b B, err error) {
	l := len(dst)
	b = dst
	b = append(b, make(B, len(src)/2)...)
	if err = xhex.Decode(b[l:], src); chk.E(err) {
		return
	}
	return
}
