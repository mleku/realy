package envelopes

import (
	"io"
)

// Marshaller is a function signature the same as the codec.JSON Marshal but
// without the requirement of there being a full implementation or declared
// receiver variable of this interface. Used here to encapsulate one or more
// other data structures into an envelope.
type Marshaller func(dst []byte) (b []byte)

// Marshal is a parser for dynamic typed arrays like nosttr codec.Envelope
// types.
func Marshal(dst []byte, label string, m Marshaller) (b []byte) {
	b = dst
	b = append(b, '[', '"')
	b = append(b, label...)
	b = append(b, '"', ',')
	b = m(b)
	b = append(b, ']')
	return
}

// SkipToTheEnd scans forward after all fields in an envelope have been read to
// find the closing bracket.
func SkipToTheEnd(dst []byte) (rem []byte, err error) {
	if len(dst) == 0 {
		return
	}
	rem = dst
	// we have everything, just need to snip the end
	for ; len(rem) > 0; rem = rem[1:] {
		if rem[0] == ']' {
			rem = rem[:0]
			return
		}
	}
	err = io.EOF
	return
}
