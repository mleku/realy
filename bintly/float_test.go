package bintly

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutFloat64(t *testing.T) {
	var useCases = []struct {
		description string
		i           float64
	}{
		{
			description: "max float64",
			i:           math.MaxFloat64,
		},
		{
			description: "min float64",
			i:           0,
		},
		{
			description: "random float64",
			i:           rand.Float64(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutFloat64(data, useCase.i)
		actual := Float64(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func TestPutFloat32(t *testing.T) {
	var useCases = []struct {
		description string
		i           float32
	}{
		{
			description: "max float32",
			i:           math.MaxFloat32,
		},
		{
			description: "min float32",
			i:           0,
		},
		{
			description: "random float32",
			i:           rand.Float32(),
		},
	}

	for _, useCase := range useCases {
		var data = make([]byte, 8)
		PutFloat32(data, useCase.i)
		actual := Float32(data)
		assert.EqualValues(t, useCase.i, actual, useCase.description)
	}

}

func BenchmarkPutFloat64(b *testing.B) {
	var bs = make([]byte, 8)
	for i := 0; i < b.N; i++ {
		for _, j := range float64Slice {
			PutFloat64(bs, j)
			a := Float64(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}

func BenchmarkPutFloat32(b *testing.B) {
	var bs = make([]byte, 8)
	for i := 0; i < b.N; i++ {
		for _, j := range float32Slice {
			PutFloat32(bs, j)
			a := Float32(bs)
			if a != j {
				assert.EqualValues(b, j, a)
			}
		}
	}
}
