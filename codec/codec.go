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
	// MarshalJSON converts the data of the type into JSON, appending it to the
	// provided slice and returning the extended slice.
	MarshalJSON(dst by) (b by, err er)
	// UnmarshalJSON decodes a JSON form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalJSON(b by) (r by, err er)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending
	// it to the provided slice.
	MarshalBinary(dst by) (b by, err er)
	// UnmarshalBinary decodes a binary form of a type back into the runtime
	// form, and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b by) (r by, err er)
}
