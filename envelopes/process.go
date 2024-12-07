package envelopes

import (
	"io"
)

type Marshaler func(dst by) (b by)

func Marshal(dst by, label string, m Marshaler) (b by) {
	b = dst
	b = append(b, '[', '"')
	b = append(b, label...)
	b = append(b, '"', ',')
	b = m(b)
	b = append(b, ']')
	return
}

func SkipToTheEnd(dst by) (rem by, err er) {
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
