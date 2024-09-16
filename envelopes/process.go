package envelopes

import (
	"io"
)

type Marshaler func(dst B) (b B, err error)

func Marshal(dst B, label string, m Marshaler) (b B, err error) {
	b = dst
	b = append(b, '[', '"')
	b = append(b, label...)
	b = append(b, '"', ',')
	if b, err = m(b); chk.E(err) {
		return
	}
	b = append(b, ']')
	return
}

func SkipToTheEnd(dst B) (rem B, err error) {
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
