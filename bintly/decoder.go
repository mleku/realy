package bintly

//Decoder represents a decoder interface
type Decoder interface {
	DecodeBinary(stream *Reader) error
}

//Allocator represents repeated type allocator
type Allocator interface {
	SetAlloc(allocation int32)
}
