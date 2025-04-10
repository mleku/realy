package json

import (
	"io"

	"realy.mleku.dev/codec"
)

// An Array is an ordered list of values. Each field is typed, to deal with the javascript
// dynamic typing. This means the array is essentially static typed, so it won't work if the
// data is not in a predictable format.
type Array struct{ V []codec.JSON }

// Marshal a list of values
func (a *Array) Marshal(dst []byte) (b []byte) {
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

// Unmarshal decodes a byte string into an array, and returns the remainder after the end of the
// array.
func (a *Array) Unmarshal(dst []byte) (rem []byte, err error) {
	rem = dst
	var openBracket bool
	var element int
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
