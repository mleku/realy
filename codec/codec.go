// Package codec is a set of interfaces for nostr messages and message elements.
package codec

import (
	"io"
)

// Envelope is an interface for the nostr "envelope" message formats, a JSON
// array with the first field an upper case string that provides type
// information, in combination with the context of the side sending it (relay or
// client).
type Envelope interface {
	// Label returns the (uppercase) string that signifies the type of message.
	Label() string
	// Write outputs the envelope to an io.Writer
	Write(w io.Writer) (err error)
	// JSON is a somewhat simplified version of the json.Marshaler/json.Unmarshaler
	// that has no error for the Marshal side of the operation.
	JSON
}

// JSON is a somewhat simplified version of the json.Marshaler/json.Unmarshaler
// that has no error for the Marshal side of the operation.
type JSON interface {
	// Marshal converts the data of the type into JSON, appending it to the provided
	// slice and returning the extended slice.
	Marshal(dst []byte) (b []byte)
	// Unmarshal decodes a JSON form of a type back into the runtime form, and
	// returns whatever remains after the type has been decoded out.
	Unmarshal(b []byte) (r []byte, err error)
}

// Binary is a similarly simplified form of the stdlib binary Marshal/Unmarshal
// interfaces. Same as JSON it does not have an error for the MarshalBinary.
type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending it to
	// the provided slice.
	MarshalBinary(dst []byte) (b []byte)
	// UnmarshalBinary decodes a binary form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b []byte) (r []byte, err error)
}
