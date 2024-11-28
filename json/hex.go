package json

import (
	"realy.lol/hex"
	"realy.lol/text"
)

// Hex is a string representing binary data encoded as hexadecimal.
type Hex struct{ V B }

func (h *Hex) Marshal(dst B) (b B) {
	b = dst
	b = append(b, '"')
	b = hex.EncAppend(b, h.V)
	b = append(b, '"')
	return
}

func (h *Hex) Unmarshal(dst B) (rem B, err E) {
	var c B
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	h.V = make(B, len(c)/2)
	var n N
	if n, err = hex.DecBytes(h.V, c); chk.E(err) {
		err = errorf.E("failed to decode hex at position %d: %s", n, err)
		return
	}
	return
}
