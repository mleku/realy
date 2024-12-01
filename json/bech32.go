package json

import (
	"realy.lol/bech32encoding"
	"realy.lol/ec/bech32"
	"realy.lol/text"
)

// Bech32 is a string encoded in bech32 format including a human-readable prefix and base32
// encoded binary data. A key difference is the HRP prefix.
//
// To create a new Bech32, load the HRP variable with the intended HRP for encoding.
//
// When decoding, point this variable at the expected HRP, if it doesn't match in what is
// encoded, it returns an error.
type Bech32 struct{ HRP, V by }

func (b2 *Bech32) Marshal(dst by) (b by) {
	var err er
	var b5 by
	if b5, err = bech32encoding.ConvertForBech32(b2.V); chk.E(err) {
		return
	}
	var bb by
	if bb, err = bech32.Encode(b2.HRP, b5); chk.E(err) {
		return
	}
	b = append(dst, '"')
	b = append(b, bb...)
	b = append(b, '"')
	return
}

func (b2 *Bech32) Unmarshal(dst by) (rem by, err er) {
	var c by
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	var b5, hrp by
	if hrp, b5, err = bech32.Decode(c); chk.E(err) {
		return
	}
	if !equals(hrp, b2.HRP) {
		err = errorf.E("invalid HRP, got '%s' expected '%s'", hrp, b2.HRP)
		return
	}
	if b2.V, err = bech32encoding.ConvertFromBech32(b5); chk.E(err) {
		return
	}
	return
}
