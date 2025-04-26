package bintly

//Encoder defines an encoder interface
type Encoder interface {
	//EncodeBinary writes data to the stream
	EncodeBinary(stream *Writer) error
}

//Alloc represents repeated type allocator
type Alloc interface {
	//Alloc returns size of repeated type
	Alloc() int32
}
