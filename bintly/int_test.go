package bintly

import (
	"encoding/binary"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutInt(t *testing.T) {
	var useCases = []struct {
		description string
		i           int
	}{
		{
			description: "max uint64",
			i:           math.MaxUint32,
		},
		{
			description: "min uint64",
			i:           0,
		},
		{
			description: "random uint64",
			i:           int(rand.Uint64()),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutInt(data, useCase.i)
		actual := Int(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetInt(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}
}

func TestPutUint(t *testing.T) {
	var useCases = []struct {
		description string
		i           uint
	}{
		{
			description: "max uint64",
			i:           math.MaxUint32,
		},
		{
			description: "min uint64",
			i:           0,
		},
		{
			description: "random uint64",
			i:           uint(rand.Uint32()),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutUint(data, useCase.i)
		actual := Uint(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetUint(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}
}

func TestPutUint64(t *testing.T) {
	var useCases = []struct {
		description string
		i           uint64
	}{
		{
			description: "max uint64",
			i:           math.MaxUint64,
		},
		{
			description: "min uint64",
			i:           0,
		},
		{
			description: "random uint64",
			i:           rand.Uint64(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutUint64(data, useCase.i)
		actual := Uint64(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetUint64(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func BenchmarkPutUint64(b *testing.B) {
	var bs = make([]byte, 8)
	for i := 0; i < b.N; i++ {
		for _, j := range uint64slice {
			PutUint64(bs, j)
			a := Uint64(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func BenchmarkPutUint64_Binary(b *testing.B) {
	coder := binary.LittleEndian
	var bs = make([]byte, 8)
	for i := 0; i < b.N; i++ {
		for _, j := range uint64slice {
			coder.PutUint64(bs, j)
			a := coder.Uint64(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func TestPutUint32(t *testing.T) {
	var useCases = []struct {
		description string
		i           uint32
	}{
		{
			description: "max uint32",
			i:           math.MaxUint32,
		},
		{
			description: "min uint32",
			i:           0,
		},
		{
			description: "random uint32",
			i:           rand.Uint32(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 4)
		PutUint32(data, useCase.i)
		actual := Uint32(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetUint32(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func BenchmarkPutUint32(b *testing.B) {
	var bs = make([]byte, 4)
	for i := 0; i < b.N; i++ {
		for _, j := range uint32Slice {
			PutUint32(bs, j)
			a := Uint32(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func BenchmarkPutUint32_Binary(b *testing.B) {
	coder := binary.LittleEndian
	var bs = make([]byte, 16)
	for i := 0; i < b.N; i++ {
		for _, j := range uint32Slice {
			coder.PutUint32(bs, j)
			a := coder.Uint32(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func TestPutUint16(t *testing.T) {
	var useCases = []struct {
		description string
		i           uint16
	}{
		{
			description: "max uint16",
			i:           math.MaxUint16,
		},
		{
			description: "min uint16",
			i:           0,
		},
		{
			description: "random uint16",
			i:           uint16(rand.Uint32() % math.MaxUint16),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 4)
		PutUint16(data, useCase.i)
		actual := Uint16(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetUint16(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func BenchmarkPutUint16(b *testing.B) {
	var bs = make([]byte, 4)
	for i := 0; i < b.N; i++ {
		for _, j := range uint16Slice {
			PutUint16(bs, j)
			a := Uint16(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func BenchmarkPutUint16_Binary(b *testing.B) {
	coder := binary.LittleEndian
	var bs = make([]byte, 16)
	for i := 0; i < b.N; i++ {
		for _, j := range uint16Slice {
			coder.PutUint16(bs, j)
			a := coder.Uint16(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func TestPutInt64(t *testing.T) {
	var useCases = []struct {
		description string
		i           int64
	}{
		{
			description: "max int64",
			i:           math.MaxInt64,
		},
		{
			description: "min int64",
			i:           0,
		},
		{
			description: "random int64",
			i:           rand.Int63(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutInt64(data, useCase.i)
		actual := Int64(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetInt64(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}
}

func Benchmark_Int64(b *testing.B) {
	var data = make([]byte, 8)
	expect := int64Slice[0]
	PutInt64(data, expect)
	for i := 0; i < b.N; i++ {
		actual := Int64(data)
		if actual != expect {
			assert.EqualValues(b, expect, actual)
		}
	}
}

func Benchmark_GetInt64(b *testing.B) {
	var data = make([]byte, 8)
	expect := int64Slice[0]
	PutInt64(data, expect)
	var actual int64
	for i := 0; i < b.N; i++ {
		GetInt64(data, &actual)
		if actual != expect {
			assert.EqualValues(b, expect, actual)
		}
	}

}

func TestPutInt32(t *testing.T) {
	var useCases = []struct {
		description string
		i           int32
	}{
		{
			description: "max int32",
			i:           math.MaxInt32,
		},
		{
			description: "min int32",
			i:           0,
		},
		{
			description: "random int32",
			i:           rand.Int31(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 4)
		PutInt32(data, useCase.i)
		actual := Int32(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetInt32(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func TestPutInt16(t *testing.T) {
	var useCases = []struct {
		description string
		i           int16
	}{
		{
			description: "max int16",
			i:           math.MaxInt16,
		},
		{
			description: "min int16",
			i:           0,
		},
		{
			description: "random int16",
			i:           int16(rand.Int31() % math.MaxInt16),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 4)
		PutInt16(data, useCase.i)
		actual := Int16(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
		actual = 0
		GetInt16(data, &actual)
		assert.EqualValues(t, useCase.i, actual, useCase.description)

	}

}
