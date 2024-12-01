package json

import (
	"realy.lol/text"
)

// String is a regular string. Must be escaped as per nip-01. Bytes stored in it are not
// escaped, only must be escaped to output JSON.
//
// Again like the other types, create a new String with:
//
//	str := json.String{}
//
// There is also a convenience NewString function that generically automatically converts actual
// golang strings to save the caller from doing so.
type String struct{ V by }

func NewString[V st | by](s V) *String { return &String{by(s)} }

func (s *String) Marshal(dst by) (b by) {
	b = text.AppendQuote(dst, s.V, text.NostrEscape)
	return
}

func (s *String) Unmarshal(dst by) (rem by, err er) {
	var c by
	if c, rem, err = text.UnmarshalQuoted(dst); chk.E(err) {
		return
	}
	s.V = c
	return
}
