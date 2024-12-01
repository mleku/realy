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

func EncAppend(dst, src by) (b by) {
	l := len(dst)
	dst = append(dst, make(by, len(src)*2)...)
	xhex.Encode(dst[l:], src)
	return dst
}

func DecAppend(dst, src by) (b by, err er) {
	l := len(dst)
	dst = append(dst, make(by, len(src)/2)...)
	err = xhex.Decode(dst[l:], src)
	return
}
