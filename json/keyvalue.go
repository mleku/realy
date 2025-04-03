package json

import (
	"io"

	"realy.lol/codec"
)

// An Object is an (not necessarily) ordered list of KeyValue.
type Object struct{ V []KeyValue }

// A KeyValue is a field in an Object.
type KeyValue struct {
	Key   []byte
	Value codec.JSON
}

// Marshal a KeyValue - ie, a JSON object key, from the provided byte string.
func (k *KeyValue) Marshal(dst []byte) (b []byte) {
	b = (&String{k.Key}).Marshal(dst)
	b = append(b, ':')
	b = k.Value.Marshal(b)
	return
}

// Unmarshal a JSON object key from a provided byte string.
func (k *KeyValue) Unmarshal(dst []byte) (rem []byte, err error) {
	rem = dst
	s := &String{}
	if rem, err = s.Unmarshal(rem); chk.E(err) {
		return
	}
	k.Key = s.V
	// note that we aren't checking there isn't gobbledygook between the sentinels...
	// this code would allow:
	//
	//   `"key"abcdefghij-literally anything:\nliterally\t \f anythingelse \r"string"`
	//
	// but once the data is in what does it matter, if it's valid. and such garbage won't
	// be part of a canonical form to generate a hash so that still works.
	for ; len(rem) >= 0; rem = rem[1:] {
		if len(rem) == 0 {
			err = io.EOF
			return
		}
		// advance to colon
		if rem[0] == ':' {
			// consume the colon, and end search
			rem = rem[1:]
			break
		}
	}
	// there should now be a value
	if rem, err = k.Value.Unmarshal(rem); chk.E(err) {
		return
	}
	return
}
