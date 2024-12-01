package text

import (
	"realy.lol/hex"
)

func AppendHexFromBinary(dst, src by, quote bo) (b by) {
	if quote {
		dst = AppendQuote(dst, src, hex.EncAppend)
	} else {
		dst = hex.EncAppend(dst, src)
	}
	b = dst
	return
}

func AppendBinaryFromHex(dst, src by, unquote bo) (b by,
	err er) {
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
