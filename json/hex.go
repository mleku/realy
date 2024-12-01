package json

import (
	"realy.lol/hex"
	"realy.lol/text"
)

// Hex is a string representing binary data encoded as hexadecimal.
type Hex struct{ V by }

func (h *Hex) Marshal(dst by) (b by) {
	b = dst
	b = append(b, '"')
	b = hex.EncAppend(b, h.V)
	b = append(b, '"')
	return
}

func (h *Hex) Unmarshal(dst by) (rem by, err er) {
	var c by
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	h.V = make(by, len(c)/2)
	var n no
	if n, err = hex.DecBytes(h.V, c); chk.E(err) {
		err = errorf.E("failed to decode hex at position %d: %s", n, err)
		return
	}
	return
}
