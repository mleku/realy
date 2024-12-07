package codec

import (
	"io"
)

type Envelope interface {
	Label() st
	Write(w io.Writer) (err er)
	JSON
}

type JSON interface {
	// Marshal converts the data of the type into JSON, appending it to the provided
	// slice and returning the extended slice.
	Marshal(dst by) (b by)
	// Unmarshal decodes a JSON form of a type back into the runtime form, and
	// returns whatever remains after the type has been decoded out.
	Unmarshal(b by) (r by, err er)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending it to
	// the provided slice.
	MarshalBinary(dst by) (b by)
	// UnmarshalBinary decodes a binary form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b by) (r by, err er)
}
