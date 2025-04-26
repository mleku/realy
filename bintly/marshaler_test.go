package bintly

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type marshStruct01 struct {
	ID          int
	OptionalID  *int
	Name        string
	Options     []string
	Elements    []int
	Coordinates []float64
}

func (m *marshStruct01) DecodeBinary(stream *Reader) error {
	stream.Int(&m.ID)
	stream.IntPtr(&m.OptionalID)
	stream.String(&m.Name)
	stream.Strings(&m.Options)
	stream.Ints(&m.Elements)
	stream.Float64s(&m.Coordinates)
	return nil
}

func (m *marshStruct01) EncodeBinary(stream *Writer) error {
	stream.Int(m.ID)
	stream.IntPtr(m.OptionalID)
	stream.String(m.Name)
	stream.Strings(m.Options)
	stream.Ints(m.Elements)
	stream.Float64s(m.Coordinates)
	return nil
}

type marshStruct02 struct {
	A1 int
	B1 *int
	C1 []int
	D1 uint
	E1 *uint
	F1 []uint

	A2 int64
	B2 *int64
	C2 []int64
	D2 uint64
	E2 *uint64
	F2 []uint64

	A3 int32
	B3 *int32
	C3 []int32
	D3 uint32
	E3 *uint32
	F3 []uint32

	A4 int16
	B4 *int16
	C4 []int16
	D4 uint16
	E4 *uint16
	F4 []uint16
	A5 int8
	B5 *int8
	C5 []int8
	D5 uint8
	E5 *uint8
	F5 []uint8
	G  bool
	H  string
}

func (m *marshStruct02) EncodeBinary(stream *Writer) error {
	stream.Int(m.A1)
	stream.IntPtr(m.B1)
	stream.Ints(m.C1)
	stream.Uint(m.D1)
	stream.UintPtr(m.E1)
	stream.Uints(m.F1)

	stream.Int64(m.A2)
	stream.Int64Ptr(m.B2)
	stream.Int64s(m.C2)
	stream.Uint64(m.D2)
	stream.Uint64Ptr(m.E2)
	stream.Uint64s(m.F2)

	stream.Int32(m.A3)
	stream.Int32Ptr(m.B3)
	stream.Int32s(m.C3)
	stream.Uint32(m.D3)
	stream.Uint32Ptr(m.E3)
	stream.Uint32s(m.F3)

	stream.Int16(m.A4)
	stream.Int16Ptr(m.B4)
	stream.Int16s(m.C4)
	stream.Uint16(m.D4)
	stream.Uint16Ptr(m.E4)
	stream.Uint16s(m.F4)

	stream.Int8(m.A5)
	stream.Int8Ptr(m.B5)
	stream.Int8s(m.C5)
	stream.Uint8(m.D5)
	stream.Uint8Ptr(m.E5)
	stream.Uint8s(m.F5)
	stream.Bool(m.G)
	stream.String(m.H)
	return nil
}

func (m *marshStruct02) DecodeBinary(stream *Reader) error {
	stream.Int(&m.A1)
	stream.IntPtr(&m.B1)
	stream.Ints(&m.C1)
	stream.Uint(&m.D1)
	stream.UintPtr(&m.E1)
	stream.Uints(&m.F1)

	stream.Int64(&m.A2)
	stream.Int64Ptr(&m.B2)
	stream.Int64s(&m.C2)
	stream.Uint64(&m.D2)
	stream.Uint64Ptr(&m.E2)
	stream.Uint64s(&m.F2)

	stream.Int32(&m.A3)
	stream.Int32Ptr(&m.B3)
	stream.Int32s(&m.C3)
	stream.Uint32(&m.D3)
	stream.Uint32Ptr(&m.E3)
	stream.Uint32s(&m.F3)

	stream.Int16(&m.A4)
	stream.Int16Ptr(&m.B4)
	stream.Int16s(&m.C4)
	stream.Uint16(&m.D4)
	stream.Uint16Ptr(&m.E4)
	stream.Uint16s(&m.F4)

	stream.Int8(&m.A5)
	stream.Int8Ptr(&m.B5)
	stream.Int8s(&m.C5)
	stream.Uint8(&m.D5)
	stream.Uint8Ptr(&m.E5)
	stream.Uint8s(&m.F5)
	stream.Bool(&m.G)
	stream.String(&m.H)
	return nil
}

func TestMarshal(t *testing.T) {

	var useCases = []struct {
		description string
		source      interface{}
	}{
		{
			description: "basic struct",
			source: &marshStruct01{
				ID:   100,
				Name: "test Me",
				Options: []string{
					"1",
					"20",
					"300",
				},
				Elements:    []int{30, 40, 50, 100, 1000, 2000},
				Coordinates: []float64{1.2, 0.2, 5.3, 6.3},
			},
		},
		{
			description: "large struct",
			source: &marshStruct02{
				A1: 101,
				B1: &intSlice[0],
				C1: intSlice,
				D1: 201,
				E1: &uintSlice[0],
				F1: uintSlice,
				A2: math.MinInt32,
				B2: &int64Slice[0],
				C2: int64Slice,
				D2: uint64slice[0],
				E2: &uint64slice[0],
				F2: uint64slice,
				A3: int32Slice[0],
				B3: &int32Slice[0],
				C3: int32Slice,
				D3: uint32Slice[0],
				E3: &uint32Slice[0],
				F3: uint32Slice,
				A4: 0,
				B4: nil,
				C4: nil,
				D4: uint16Slice[0],
				E4: &uint16Slice[0],
				F4: uint16Slice,
				A5: int8Slice[0],
				B5: &int8Slice[0],
				C5: int8Slice,
				D5: uint8Slice[0],
				E5: &uint8Slice[0],
				F5: uint8Slice,
				G:  false,
				H:  "test me",
			},
		},
	}

	for _, useCase := range useCases {
		data, err := Marshal(useCase.source)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		var actual interface{}
		vType := reflect.TypeOf(useCase.source)
		if vType.Kind() == reflect.Ptr {
			vType = vType.Elem()
		}
		actual = reflect.New(vType).Interface()
		err = Unmarshal(data, actual)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.source, actual, useCase.description)
	}
}
