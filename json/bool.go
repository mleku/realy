package json

import (
	"io"
)

// Bool can be either `true` or `false` and only lower case. Initialize these values as follows:
//
//	truthValue := &json.Bool{}
//
// to get a default value which is false.
type Bool struct{ V bool }

const T = "true"
const F = "false"

var Bools = map[bool]B{
	true:  B(T),
	false: B(F),
}

func (b2 *Bool) Marshal(dst B) (b B) {
	b = dst
	b = append(b, Bools[b2.V]...)
	return
}

func (b2 *Bool) Unmarshal(dst B) (rem B, err E) {
	rem = dst
	// this is a shortcut evaluation because any text not in quotes in JSON is invalid so if
	// it is something other than the exact correct, the next value will not match and the
	// larger structure being unmarshalled will fail with an error.
	if rem[0] == Bools[true][0] {
		if len(rem) < len(T) {
			err = io.EOF
			return
		}
		if equals(Bools[true], rem[:len(T)]) {
			b2.V = true
			rem = rem[len(T):]
			return
		}
	}
	if rem[0] == Bools[false][0] {
		if len(rem) < len(F) {
			err = io.EOF
			return
		}
		if equals(Bools[false], rem[:len(F)]) {
			b2.V = false
			rem = rem[len(F):]
			return
		}
	}
	// if a truth value is not found in the string it will run to the end
	err = io.EOF
	return
}
