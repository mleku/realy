package json

import (
	"bytes"
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

var Bools = map[bool][]byte{
	true:  []byte(T),
	false: []byte(F),
}

// Marshal a Bool into JSON text (ie true/false)
func (b2 *Bool) Marshal(dst []byte) (b []byte) {
	b = dst
	b = append(b, Bools[b2.V]...)
	return
}

// Unmarshal a byte string that should be containing a boolean true/false.
//
// this is a shortcut evaluation because any text not in quotes in JSON is invalid so if
// it is something other than the exact correct, the next value will not match and the
// larger structure being unmarshalled will fail with an error.
func (b2 *Bool) Unmarshal(dst []byte) (rem []byte, err error) {
	rem = dst
	if rem[0] == Bools[true][0] {
		if len(rem) < len(T) {
			err = io.EOF
			return
		}
		if bytes.Equal(Bools[true], rem[:len(T)]) {
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
		if bytes.Equal(Bools[false], rem[:len(F)]) {
			b2.V = false
			rem = rem[len(F):]
			return
		}
	}
	// if a truth value is not found in the string it will run to the end
	err = io.EOF
	return
}
