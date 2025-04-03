package json

import (
	"realy.lol/hex"
	"realy.lol/text"
)

// Hex is a string representing binary data encoded as hexadecimal.
type Hex struct{ V []byte }

// Marshal a byte string into hexadecimal wrapped in double quotes.
func (h *Hex) Marshal(dst []byte) (b []byte) {
	b = dst
	b = append(b, '"')
	b = hex.EncAppend(b, h.V)
	b = append(b, '"')
	return
}

// Unmarshal a string wrapped in double quotes that should be a hexadecimal string. If it fails,
// it will return an error.
func (h *Hex) Unmarshal(dst []byte) (rem []byte, err error) {
	var c []byte
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	h.V = make([]byte, len(c)/2)
	var n int
	if n, err = hex.DecBytes(h.V, c); chk.E(err) {
		err = errorf.E("failed to decode hex at position %d: %s", n, err)
		return
	}
	return
}
