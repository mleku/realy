package bintly

import (
	"encoding/binary"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PutInts(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		ints        []int
	}{
		{
			description: "tiny slice",
			ints:        []int{math.MaxInt32, 0, 10, math.MaxUint16, math.MaxUint32, math.MaxUint8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "medium slice 2",
			size:        100,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.ints) == 0 {
			useCase.ints = make([]int, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.ints[i] = int(rand.Int31())
			}
		}
		var data = make([]byte, 8*len(useCase.ints))
		PutInts(data, useCase.ints)
		actual := Ints(data)
		assert.EqualValues(t, useCase.ints, actual, useCase.description)
	}
}

func Test_PutUint64s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		uints       []uint64
	}{
		{
			description: "tiny slice",
			uints:       []uint64{math.MaxUint64, 0, 10, math.MaxUint16, math.MaxUint32, math.MaxUint8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.uints) == 0 {
			useCase.uints = make([]uint64, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.uints[i] = rand.Uint64()
			}
		}
		var data = make([]byte, 8*len(useCase.uints))
		PutUint64s(data, useCase.uints)
		actual := Uint64s(data)
		assert.EqualValues(t, useCase.uints, actual, useCase.description)
	}
}

func BenchmarkPutUint64s(b *testing.B) {
	var data = make([]byte, len(uint64slice)*8)
	for i := 0; i < b.N; i++ {
		PutUint64s(data, uint64slice)
		actual := Uint64s(data)
		for j, v := range uint64slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkPutUint64s_Binary(b *testing.B) {
	var data = make([]byte, len(uint64slice)*8)
	coder := binary.LittleEndian
	for i := 0; i < b.N; i++ {
		for j, v := range uint64slice {
			coder.PutUint64(data[j*8:], v)
		}
		var actual = make([]uint64, len(uint64slice))
		for j := range uint64slice {
			actual[j] = coder.Uint64(data[j*8:])
		}
		for j, v := range uint64slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkGetUint64s(b *testing.B) {
	var data = make([]byte, len(uint64slice)*8)
	var actual = make([]uint64, len(uint64slice))
	for i := 0; i < b.N; i++ {
		PutUint64s(data, uint64slice)
		GetUint64s(data, actual)
		for j, v := range uint64slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkGetUint64s_Binary(b *testing.B) {
	var data = make([]byte, len(uint64slice)*8)
	coder := binary.LittleEndian
	var actual = make([]uint64, len(uint64slice))
	for i := 0; i < b.N; i++ {
		for j, v := range uint64slice {
			coder.PutUint64(data[j*8:], v)
		}
		for j := range uint64slice {
			actual[j] = coder.Uint64(data[j*8:])
		}
		for j, v := range uint64slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func Test_PutUint32s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		uints       []uint32
	}{
		{
			description: "tiny slice",
			uints:       []uint32{math.MaxUint32, 0, 10, math.MaxUint16, math.MaxUint32, math.MaxUint8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.uints) == 0 {
			useCase.uints = make([]uint32, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.uints[i] = rand.Uint32()
			}
		}
		var data = make([]byte, 4*len(useCase.uints))
		PutUint32s(data, useCase.uints)
		actual := Uint32s(data)
		assert.EqualValues(t, useCase.uints, actual, useCase.description)
	}
}

func BenchmarkPutUint32s(b *testing.B) {
	var data = make([]byte, len(uint32Slice)*4)
	for i := 0; i < b.N; i++ {
		PutUint32s(data, uint32Slice)
		actual := Uint32s(data)
		for j, v := range uint32Slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkPutUint32s_Binary(b *testing.B) {
	var data = make([]byte, len(uint32Slice)*4)
	coder := binary.LittleEndian
	for i := 0; i < b.N; i++ {
		for j, v := range uint32Slice {
			coder.PutUint32(data[j*4:], v)
		}
		var actual = make([]uint32, len(uint32Slice))
		for j := range uint32Slice {
			actual[j] = coder.Uint32(data[j*4:])
		}
		for j, v := range uint32Slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkGetUint32s(b *testing.B) {
	var data = make([]byte, len(uint32Slice)*4)
	var actual = make([]uint32, len(uint32Slice))
	for i := 0; i < b.N; i++ {
		PutUint32s(data, uint32Slice)
		GetUint32s(data, actual)
		for j, v := range uint32Slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func BenchmarkGetUint32s_Binary(b *testing.B) {
	var data = make([]byte, len(uint32Slice)*8)
	coder := binary.LittleEndian
	var actual = make([]uint32, len(uint32Slice))
	for i := 0; i < b.N; i++ {
		for j, v := range uint32Slice {
			coder.PutUint32(data[j*4:], v)
		}
		for j := range uint32Slice {
			actual[j] = coder.Uint32(data[j*4:])
		}
		for j, v := range uint32Slice {
			if v != actual[j] {
				assert.EqualValues(b, v, actual[j])
			}
		}
	}
}

func Test_PutUint16s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		uints       []uint16
	}{
		{
			description: "tiny slice",
			uints:       []uint16{math.MaxUint16, 0, 10, math.MaxUint16, math.MaxUint16, math.MaxUint8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.uints) == 0 {
			useCase.uints = make([]uint16, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.uints[i] = uint16(rand.Uint32() % math.MaxUint16)
			}
		}
		var data = make([]byte, 2*len(useCase.uints))
		PutUint16s(data, useCase.uints)
		actual := Uint16s(data)
		assert.EqualValues(t, useCase.uints, actual, useCase.description)
	}
}

func Test_PutInt64s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		ints        []int64
	}{
		{
			description: "tiny slice",
			ints:        []int64{math.MaxInt64, 0, 10, math.MaxInt16, math.MaxInt32, math.MaxInt8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.ints) == 0 {
			useCase.ints = make([]int64, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.ints[i] = rand.Int63()
			}
		}
		var data = make([]byte, 8*len(useCase.ints))
		PutInt64s(data, useCase.ints)
		actual := Int64s(data)
		assert.EqualValues(t, useCase.ints, actual, useCase.description)
	}
}

func Test_PutInt32s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		ints        []int32
	}{
		{
			description: "tiny slice",
			ints:        []int32{math.MaxInt32, 0, 10, math.MaxInt16, math.MaxInt32, math.MaxInt8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.ints) == 0 {
			useCase.ints = make([]int32, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.ints[i] = rand.Int31()
			}
		}
		var data = make([]byte, 4*len(useCase.ints))
		PutInt32s(data, useCase.ints)
		actual := Int32s(data)
		assert.EqualValues(t, useCase.ints, actual, useCase.description)
	}
}

func Test_PutInt16s(t *testing.T) {
	var useCases = []struct {
		description string
		size        int
		ints        []int16
	}{
		{
			description: "tiny slice",
			ints:        []int16{math.MaxInt16, 0, 10, math.MaxInt16, math.MaxInt16, math.MaxInt8},
		},
		{
			description: "small slice",
			size:        257,
		},
		{
			description: "medium slice",
			size:        3 * 1024,
		},
		{
			description: "large slice",
			size:        1024 * 128,
		},
	}

	for _, useCase := range useCases {
		if len(useCase.ints) == 0 {
			useCase.ints = make([]int16, useCase.size)
			for i := 0; i < useCase.size; i++ {
				useCase.ints[i] = int16(rand.Int31() % math.MaxInt16)
			}
		}
		var data = make([]byte, 2*len(useCase.ints))
		PutInt16s(data, useCase.ints)
		actual := Int16s(data)
		assert.EqualValues(t, useCase.ints, actual, useCase.description)
	}
}
