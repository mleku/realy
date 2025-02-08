package codec

import (
	"io"
)

type Envelope interface {
	Label() string
	Write(w io.Writer) (err error)
	JSON
}

type JSON interface {
	// Marshal converts the data of the type into JSON, appending it to the provided
	// slice and returning the extended slice.
	Marshal(dst []byte) (b []byte)
	// Unmarshal decodes a JSON form of a type back into the runtime form, and
	// returns whatever remains after the type has been decoded out.
	Unmarshal(b []byte) (r []byte, err error)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending it to
	// the provided slice.
	MarshalBinary(dst []byte) (b []byte)
	// UnmarshalBinary decodes a binary form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b []byte) (r []byte, err error)
}
