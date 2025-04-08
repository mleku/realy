package text

import (
	"realy.lol/hex"
)

// AppendHexFromBinary appends to a hex output from binary input.
func AppendHexFromBinary(dst, src []byte, quote bool) (b []byte) {
	if quote {
		dst = AppendQuote(dst, src, hex.EncAppend)
	} else {
		dst = hex.EncAppend(dst, src)
	}
	b = dst
	return
}

// AppendBinaryFromHex encodes binary input as hex and appends it to the output.
func AppendBinaryFromHex(dst, src []byte, unquote bool) (b []byte,
	err error) {
	if unquote {
		if dst, err = hex.DecAppend(dst,
			Unquote(src)); chk.E(err) {
			return
		}
	} else {
		if dst, err = hex.DecAppend(dst, src); chk.E(err) {
			return
		}
	}
	b = dst
	return
}
