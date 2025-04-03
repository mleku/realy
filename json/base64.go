package json

import (
	"bytes"
	"encoding/base64"

	"realy.lol/text"
)

// Base64 is a string representing binary data encoded as base64.
type Base64 struct{ V []byte }

// Marshal encodes a byte string into base64. This uses standard encoding, not URL encoding.
func (b2 *Base64) Marshal(dst []byte) (b []byte) {
	b = dst
	buf := &bytes.Buffer{}
	b = append(b, '"')
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	var err error
	if _, err = enc.Write(b2.V); chk.E(err) {
		return
	}
	b = append(b, buf.Bytes()...)
	b = append(b, '"')
	return
}

// Unmarshal a base64 standard encoded string into a byte string.
func (b2 *Base64) Unmarshal(dst []byte) (rem []byte, err error) {
	var c []byte
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	var n int
	bb := make([]byte, len(c)*6/8)
	if n, err = base64.StdEncoding.Decode(bb, c); chk.E(err) {
		err = errorf.E("failed to decode base64 at position %d: %s", n, err)
		return
	}
	b2.V = bb
	return
}
