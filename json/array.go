package json

import (
	"io"
	"realy.lol/codec"
)

// An Array is an ordered list of values.
type Array struct{ V []codec.JSON }

func (a *Array) Marshal(dst by) (b by) {
	b = dst
	b = append(b, '[')
	last := len(a.V) - 1
	for i, v := range a.V {
		b = v.Marshal(b)
		if i != last {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (a *Array) Unmarshal(dst by) (rem by, err er) {
	rem = dst
	var openBracket bo
	var element no
	for ; len(rem) > 0; rem = rem[1:] {
		if !openBracket && rem[0] == '[' {
			openBracket = true
			continue
		}
		if openBracket {
			if rem[0] == ',' {
				continue
			} else if rem[0] == ']' {
				rem = rem[1:]
				return
			}
			// element marshallers already know to skip until the known sign of the beginning of
			// their content, eg quotes, numerical value, etc.
			if rem, err = a.V[element].Unmarshal(rem); chk.E(err) {
				return
			}
			element++
			if len(rem) < 1 {
				err = io.EOF
				return
			}
			if rem[0] == ']' {
				rem = rem[1:]
				// done
				return
			}
			if element == len(a.V) {
				err = io.EOF
				return
			}
		}
	}
	return
}
