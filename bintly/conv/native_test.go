package conv

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNative(t *testing.T) {
	type intAlias int
	type uintAlias uint
	type int64Alias int64
	type uint64Alias uint64
	type int32Alias int32
	type uint32Alias uint32
	type int16Alias int16
	type uint16Alias uint16
	type int8Alias int8
	type uint8Alias uint8
	type float64Alias float64
	type float32Alias float32
	type boolAlias bool
	type stringAlias string
	type btesType []byte

	var useCases = []struct {
		description string
		value       interface{}
		nativeType  *reflect.Type
	}{
		{
			description: "int alias",
			value:       intAlias(0),
			nativeType:  &intType,
		},
		{
			description: "uint alias",
			value:       uintAlias(0),
			nativeType:  &uintType,
		},
		{
			description: "int64 alias",
			value:       int64Alias(0),
			nativeType:  &int64Type,
		},
		{
			description: "uint64 alias",
			value:       uint64Alias(0),
			nativeType:  &uint64Type,
		},
		{
			description: "int32 alias",
			value:       int32Alias(0),
			nativeType:  &int32Type,
		},
		{
			description: "uint32 alias",
			value:       uint32Alias(0),
			nativeType:  &uint32Type,
		},
		{
			description: "int16 alias",
			value:       int16Alias(0),
			nativeType:  &int16Type,
		},
		{
			description: "uint16 alias",
			value:       uint16Alias(0),
			nativeType:  &uint16Type,
		},
		{
			description: "int8 alias",
			value:       int8Alias(0),
			nativeType:  &int8Type,
		},
		{
			description: "uint8 alias",
			value:       uint8Alias(0),
			nativeType:  &uint8Type,
		},
		{
			description: "bool alias",
			value:       boolAlias(true),
			nativeType:  &boolType,
		},
		{
			description: "float64 alias",
			value:       float64Alias(0.1),
			nativeType:  &float64Type,
		},
		{
			description: "float32 alias",
			value:       float32Alias(0.1),
			nativeType:  &float32Type,
		},
		{
			description: "string alias",
			value:       stringAlias("a"),
			nativeType:  &stringType,
		},
		{
			description: "[]byte alias",
			value:       btesType("a"),
			nativeType:  &bytesType,
		},
	}

	for _, useCase := range useCases {
		assert.False(t, IsNative(reflect.TypeOf(useCase.value)), useCase.description)
		assert.True(t, IsNative(*useCase.nativeType), useCase.description)

		actual := *MatchNative(reflect.TypeOf(useCase.value))
		expect := *useCase.nativeType
		assert.EqualValues(t, expect.Name(), actual.Name(), useCase.description)
	}

}
