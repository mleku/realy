package json

import (
	"bytes"

	"realy.mleku.dev/bech32encoding"
	"realy.mleku.dev/ec/bech32"
	"realy.mleku.dev/text"
)

// Bech32 is a string encoded in bech32 format including a human-readable prefix and base32
// encoded binary data. A key difference is the HRP prefix.
//
// To create a new Bech32, load the HRP variable with the intended HRP for encoding.
//
// When decoding, point this variable at the expected HRP, if it doesn't match in what is
// encoded, it returns an error.
type Bech32 struct{ HRP, V []byte }

// Marshal a byte slice, with a given HRP prefix into a Bech32 string.
func (b2 *Bech32) Marshal(dst []byte) (b []byte) {
	var err error
	var b5 []byte
	if b5, err = bech32encoding.ConvertForBech32(b2.V); chk.E(err) {
		return
	}
	var bb []byte
	if bb, err = bech32.Encode(b2.HRP, b5); chk.E(err) {
		return
	}
	b = append(dst, '"')
	b = append(b, bb...)
	b = append(b, '"')
	return
}

// Unmarshal a Bech32 string into raw bytes, and extract the HRP prefix.
func (b2 *Bech32) Unmarshal(dst []byte) (rem []byte, err error) {
	var c []byte
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	var b5, hrp []byte
	if hrp, b5, err = bech32.Decode(c); chk.E(err) {
		return
	}
	if !bytes.Equal(hrp, b2.HRP) {
		err = errorf.E("invalid HRP, got '%s' expected '%s'", hrp, b2.HRP)
		return
	}
	if b2.V, err = bech32encoding.ConvertFromBech32(b5); chk.E(err) {
		return
	}
	return
}
