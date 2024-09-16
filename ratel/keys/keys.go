// Package keys is a composable framework for constructing badger keys from
// fields of events.
package keys

import (
	"bytes"
)

// Element is an enveloper for a type that can Read and Write its binary form.
type Element interface {
	// Write the binary form of the field into the given bytes.Buffer.
	Write(buf *bytes.Buffer)
	// Read accepts a bytes.Buffer and decodes a field from it.
	Read(buf *bytes.Buffer) Element
	// Len gives the length of the bytes output by the type.
	Len() int
}

// Write the contents of each Element to a byte slice.
func Write(elems ...Element) []byte {
	// get the length of the buffer required
	var length int
	for _, el := range elems {
		length += el.Len()
	}
	buf := bytes.NewBuffer(make([]byte, 0, length))
	// write out the data from each element
	for _, el := range elems {
		el.Write(buf)
	}
	return buf.Bytes()
}

// Read the contents of a byte slice into the provided list of Element types.
func Read(b []byte, elems ...Element) {
	buf := bytes.NewBuffer(b)
	for _, el := range elems {
		el.Read(buf)
	}
}

// Make is a convenience method to wrap a list of Element into a slice.
func Make(elems ...Element) []Element { return elems }
