package bintly

var readers = NewReaders()

//Unmarshal converts []byte to e pointer or error
func Unmarshal(data []byte, v interface{}) error {
	stream := readers.Get()
	defer readers.Put(stream)
	return UnmarshalStream(stream, data, v)
}

//UnmarshalStream converts []byte to e pointer or error
func UnmarshalStream(stream *Reader, data []byte, v interface{}) error {
	err := stream.FromBytes(data)
	if err != nil {
		return err
	}
	err = stream.Any(v)
	if err != nil {
		return err
	}

	return nil
}




//Decode converts []byte to decoder or error
func Decode(data []byte, coder Decoder) error {
	stream := readers.Get()
	err := stream.FromBytes(data)
	if err != nil {
		return err
	}
	if err = stream.Coder(coder);err == nil {
		readers.Put(stream)
	}
	return err
}