package stress

import (
	"math"
	"math/rand"
	"strings"
	"time"
)

var testSize = 80
var intSlice = make([]int, testSize)
var uintSlice = make([]uint, testSize)
var int64Slice = make([]int64, testSize)
var uint64slice = make([]uint64, testSize)
var uint32Slice = make([]uint32, testSize)
var uint16Slice = make([]uint16, testSize)
var int32Slice = make([]int32, testSize)
var int16Slice = make([]int16, testSize)
var uint8Slice = make([]uint8, testSize)
var int8Slice = make([]int8, testSize)
var float64Slice = make([]float64, testSize)
var float32Slice = make([]float32, testSize)
var stringSlice = make([]string, testSize)

func init() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(intSlice); i++ {
		intSlice[i] = int(rand.Int31())
	}
	for i := 0; i < len(uint64slice); i++ {
		uintSlice[i] = uint(rand.Uint32())
	}
	for i := 0; i < len(uint64slice); i++ {
		int64Slice[i] = rand.Int63()
	}
	for i := 0; i < len(uint64slice); i++ {
		uint64slice[i] = rand.Uint64()
	}
	for i := 0; i < len(uint32Slice); i++ {
		uint32Slice[i] = rand.Uint32()
	}
	for i := 0; i < len(uint32Slice); i++ {
		int32Slice[i] = int32(rand.Uint32())
	}
	for i := 0; i < len(uint16Slice); i++ {
		uint16Slice[i] = uint16(rand.Uint32() % math.MaxUint16)
	}
	for i := 0; i < len(int16Slice); i++ {
		int16Slice[i] = int16(rand.Uint32() % math.MaxUint16)
	}
	for i := 0; i < len(uint8Slice); i++ {
		uint8Slice[i] = uint8(rand.Uint32() % math.MaxInt8)
	}
	for i := 0; i < len(int8Slice); i++ {
		int8Slice[i] = int8(rand.Uint32() % math.MaxInt8)
	}
	for i := 0; i < len(float32Slice); i++ {
		float32Slice[i] = rand.Float32()
	}
	for i := 0; i < len(float64Slice); i++ {
		float64Slice[i] = float64(rand.Float32())
	}
	for i := 0; i < len(stringSlice); i++ {
		stringSlice[i] = strings.Repeat(string([]byte{33 + uint8(rand.Uint32()%10)}), 1+int(rand.Uint32()%10))
	}
}
