package binary

import (
	"encoding/binary"
	"math"
)

//Reader represents binary reader
type Reader struct {
	data   []byte
	offset int
	coder  binary.ByteOrder
}

//Alloc returns allocation size
func (b *Reader) Alloc() uint32 {
	alloc := b.coder.Uint32(b.data[b.offset:])
	b.offset += 4
	return alloc
}

//Int returns int
func (b *Reader) Int() int {
	v := b.coder.Uint64(b.data[b.offset:])
	b.offset += 8
	return int(v)
}

//Ints returns []int
func (b *Reader) Ints() []int {
	size := b.Alloc()
	var result = make([]int, size)
	for i := 0; i < int(size); i++ {
		result[i] = b.Int()
	}
	return result
}

//Int32 writes an int32
func (b *Reader) Int32() int32 {
	v := b.coder.Uint32(b.data[b.offset:])
	b.offset += 4
	return int32(v)
}

//Bool return bool
func (b *Reader) Bool() bool {
	result := b.data[b.offset] == 1
	b.offset++
	return result
}

//String returns string
func (b *Reader) String() string {
	size := int(b.Alloc())
	result := string(b.data[b.offset : b.offset+size])
	b.offset += size
	return result
}

//Bytes returns bytes
func (b *Reader) Bytes() []byte {
	size := int(b.Alloc())
	result := b.data[b.offset : b.offset+size]
	b.offset += size
	return result
}

//Strings returns []string
func (b *Reader) Strings() []string {
	size := int(b.Alloc())
	var result = make([]string, size)
	for i := 0; i < size; i++ {
		result[i] = b.String()
	}
	return result
}

//Float64 returns float64
func (b *Reader) Float64() float64 {
	bits := uint64(b.Int())
	return math.Float64frombits(bits)
}

//Float64s returns []float64
func (b *Reader) Float64s() []float64 {
	size := int(b.Alloc())
	var result = make([]float64, size)
	for i := 0; i < size; i++ {
		result[i] = b.Float64()
	}
	return result
}

//NewReader creates a reader
func NewReader(data []byte, coder binary.ByteOrder) *Reader {
	return &Reader{
		data:  data,
		coder: coder,
	}
}
