package stress

import (
	"encoding/binary"

	"github.com/viant/bintly"
	bin "github.com/viant/bintly/binary"
)

type (
	//BenchStruct represents a bench struct
	BenchStruct struct {
		A1 int
		A2 string
		A3 bool
		A4 float64
		A5 []int
		A6 []string
		A7 []float64
		A8 []byte
	}
	//BenchStructAlias alias to benchmark reflection
	BenchStructAlias BenchStruct
)

//DecodeBinary decode bindly stream
func (m *BenchStruct) DecodeBinary(stream *bintly.Reader) error {
	stream.Int(&m.A1)
	stream.String(&m.A2)
	stream.Bool(&m.A3)
	stream.Float64(&m.A4)
	stream.Ints(&m.A5)
	stream.Strings(&m.A6)
	stream.Float64s(&m.A7)
	stream.Uint8s(&m.A8)
	return nil
}

//EncodeBinary encodes bintly stream
func (m *BenchStruct) EncodeBinary(stream *bintly.Writer) error {
	stream.Int(m.A1)
	stream.String(m.A2)
	stream.Bool(m.A3)
	stream.Float64(m.A4)
	stream.Ints(m.A5)
	stream.Strings(m.A6)
	stream.Float64s(m.A7)
	stream.Uint8s(m.A8)
	return nil
}

//ToBytes converts to bytes with wrapped encoding/binary ByteOrder
func (m *BenchStruct) ToBytes() ([]byte, error) {
	writer := bin.NewWriter(binary.LittleEndian)
	if err := writer.Int(m.A1); err != nil {
		return nil, err
	}
	if err := writer.String(m.A2); err != nil {
		return nil, err
	}
	if err := writer.Bool(m.A3); err != nil {
		return nil, err
	}
	if err := writer.Float64(m.A4); err != nil {
		return nil, err
	}
	if err := writer.Ints(m.A5); err != nil {
		return nil, err
	}
	if err := writer.Strings(m.A6); err != nil {
		return nil, err
	}
	if err := writer.Float64s(m.A7); err != nil {
		return nil, err
	}
	if err := writer.Bytes(m.A8); err != nil {
		return nil, err
	}
	return writer.ToBytes(), nil
}

//ToBytes converts from bytes with wrapped encoding/binary ByteOrder
func (m *BenchStruct) FromBytes(bs []byte) {
	reader := bin.NewReader(bs, binary.LittleEndian)
	m.A1 = reader.Int()
	m.A2 = reader.String()
	m.A3 = reader.Bool()
	m.A4 = reader.Float64()
	m.A5 = reader.Ints()
	m.A6 = reader.Strings()
	m.A7 = reader.Float64s()
	m.A8 = reader.Bytes()
}
