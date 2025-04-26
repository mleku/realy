package bintly

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCodec_Put(t *testing.T) {
	type myInt int
	mI := myInt(333)
	var tBool = true
	var text = "testMe"
	var ts = time.Now()
	var useCases = []struct {
		description string
		value       interface{}
	}{
		{
			description: "int type",
			value:       (1024 * 1024 * 512),
		},
		{
			description: "int ptr type",
			value:       &intSlice[0],
		},
		{
			description: "[]int type",
			value:       intSlice,
		},

		{
			description: "uint type",
			value:       uint(1024 * 1024 * 512),
		},
		{
			description: "uint ptr type",
			value:       &uintSlice[0],
		},
		{
			description: "[]uint  type",
			value:       uintSlice,
		},
		{
			description: "int64 type",
			value:       int64(1024 * 1024 * 512),
		},
		{
			description: "int64 ptr type",
			value:       &int64Slice[0],
		},
		{
			description: "[]int64  type",
			value:       int64Slice,
		},
		{
			description: "uint64 type",
			value:       uint64(1024 * 1024 * 512),
		},
		{
			description: "int64 ptr type",
			value:       &uint64slice[0],
		},
		{
			description: "[]int64  type",
			value:       uint64slice,
		},
		{
			description: "int32 type",
			value:       int32(1024 * 1024 * 512),
		},
		{
			description: "*int32 type",
			value:       &int32Slice[0],
		},
		{
			description: "[]int32 type",
			value:       int32Slice,
		},
		{
			description: "uint32 type",
			value:       uint32(1024 * 1024 * 512),
		},
		{
			description: "*uint32 type",
			value:       &uint32Slice[0],
		},
		{
			description: "[]uint32 type",
			value:       uint32Slice,
		},

		{
			description: "uint16 type",
			value:       uint16(math.MaxInt16) - 100,
		},
		{
			description: "*uint16 type",
			value:       &uint16Slice[0],
		},
		{
			description: "[]uint16 type",
			value:       uint16Slice,
		},
		{
			description: "int16 type",
			value:       int16(math.MaxInt16) - 100,
		},
		{
			description: "*int16 type",
			value:       &int16Slice[0],
		},
		{
			description: "[]int16 type",
			value:       int16Slice,
		},
		{
			description: "uint8 type",
			value:       uint8(math.MaxInt8) - 100,
		},
		{
			description: "*uint8 type",
			value:       &uint8Slice[0],
		},
		{
			description: "[]uint8 type",
			value:       uint8Slice,
		},
		{
			description: "int8 type",
			value:       int8(math.MaxInt8) - 100,
		},
		{
			description: "*int8 type",
			value:       &int8Slice[0],
		},
		{
			description: "[]int8 type",
			value:       int8Slice,
		},

		{
			description: "float64 type",
			value:       float64(3.123),
		},
		{
			description: "*float64 type",
			value:       &float64Slice[0],
		},
		{
			description: "[]float64 type",
			value:       float64Slice,
		},
		{
			description: "float32 type",
			value:       float32(3.333333),
		},
		{
			description: "*float32 type",
			value:       &float32Slice[0],
		},
		{
			description: "[]float32 type",
			value:       float32Slice,
		},
		{
			description: "bool type",
			value:       true,
		},
		{
			description: "*bool type",
			value:       &tBool,
		},
		{
			description: "[]bool type",
			value:       []bool{true, false, true},
		},
		{
			description: "[]byte type",
			value:       []byte(uint8Slice),
		},
		{
			description: "string type",
			value:       "abc",
		},
		{
			description: "*string type",
			value:       &text,
		},

		{
			description: "[]string type",
			value:       stringSlice,
		},
		{
			description: "time.Time",
			value:       ts,
		},
		{
			description: "*time.Time",
			value:       &ts,
		},
		{
			description: "myInt native type",
			value:       myInt(101),
		},
		{
			description: "*myInt native type",
			value:       &mI,
		},
	}

	for _, useCase := range useCases {
		encoder := &Writer{}

		err := encoder.Any(useCase.value)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		data := encoder.Bytes()
		decoder := &Reader{}
		decoder.FromBytes(data)
		actualPtr := reflect.New(reflect.TypeOf(useCase.value))
		err = decoder.Any(actualPtr.Interface())
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actual := actualPtr.Elem().Interface()
		expect := useCase.value
		if actualPtr.Elem().Kind() == reflect.Ptr {
			actual = actualPtr.Elem().Elem().Interface()
			expect = reflect.ValueOf(expect).Elem().Interface()
		}
		if actualTime, ok := actual.(time.Time); ok {
			expectedTime := expect.(time.Time)
			actual = actualTime.UnixNano()
			expect = expectedTime.UnixNano()
		} else if actualTime, ok := actual.(*time.Time); ok {
			expectedTime := expect.(*time.Time)
			actual = actualTime.UnixNano()
			expect = expectedTime.UnixNano()
		}
		assert.EqualValues(t, expect, actual, useCase.description)
	}

}
