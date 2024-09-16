package realy

type JSON interface {
	// MarshalJSON converts the data of the type into JSON, appending it to the
	// provided slice and returning the extended slice.
	MarshalJSON(dst B) (b B, err E)
	// UnmarshalJSON decodes a JSON form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalJSON(b B) (r B, err E)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending
	// it to the provided slice.
	MarshalBinary(dst B) (b B, err E)
	// UnmarshalBinary decodes a binary form of a type back into the runtime
	// form, and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b B) (r B, err E)
}
