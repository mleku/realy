package json

import (
	"bytes"
	"encoding/base64"

	"realy.lol/text"
)

// Base64 is a string representing binary data encoded as base64.
type Base64 struct{ V B }

func (b2 *Base64) Marshal(dst B) (b B) {
	b = dst
	buf := &bytes.Buffer{}
	b = append(b, '"')
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	var err E
	if _, err = enc.Write(b2.V); chk.E(err) {
		return
	}
	b = append(b, buf.Bytes()...)
	b = append(b, '"')
	return
}

func (b2 *Base64) Unmarshal(dst B) (rem B, err E) {
	var c B
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	var n N
	bb := make(B, len(c)*6/8)
	if n, err = base64.StdEncoding.Decode(bb, c); chk.E(err) {
		err = errorf.E("failed to decode base64 at position %d: %s", n, err)
		return
	}
	b2.V = bb
	return
}
