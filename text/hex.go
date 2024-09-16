package text

import (
	"mleku.dev/hex"
)

func AppendHexFromBinary(dst, src B, quote bool) (b B) {
	if quote {
		dst = AppendQuote(dst, src, hex.EncAppend)
	} else {
		dst = hex.EncAppend(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src B, unquote bool) (b B,
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
