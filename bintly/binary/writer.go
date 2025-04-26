package binary

import (
	"bytes"
	"encoding/binary"
	"math"
)

//Writer represents binary writer
type Writer struct {
	bytes.Buffer
	buf   []byte
	coder binary.ByteOrder
}

//Alloc returns allocated size
func (b *Writer) Alloc(size uint32) error {
	b.coder.PutUint32(b.buf, size)
	_, err := b.Write(b.buf[:4])
	return err
}

//Int writes an int
func (b *Writer) Int(v int) error {
	b.coder.PutUint64(b.buf, uint64(v))
	_, err := b.Write(b.buf[:8])
	return err
}

//Int32 writes an int32
func (b *Writer) Int32(v int32) error {
	b.coder.PutUint32(b.buf, uint32(v))
	_, err := b.Write(b.buf[:4])
	return err
}

//Bool writes byte
func (b *Writer) Bool(v bool) error {
	bt := []byte{0}
	if v {
		bt[0] = 1
	}
	_, err := b.Write(bt)
	return err
}

//Ints writes []int
func (b *Writer) Ints(v []int) error {
	if err := b.Alloc(uint32(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		if err := b.Int(i); err != nil {
			return err
		}
	}
	return nil
}

//String writes string v
func (b *Writer) String(v string) error {
	if err := b.Alloc(uint32(len(v))); err != nil {
		return err
	}
	_, err := b.Write([]byte(v))
	return err
}

//Strings writes []string
func (b *Writer) Strings(vs []string) error {
	if err := b.Alloc(uint32(len(vs))); err != nil {
		return err
	}
	for _, s := range vs {
		if err := b.String(s); err != nil {
			return err
		}
	}
	return nil
}

func (b *Writer) ToBytes() []byte {
	return b.Buffer.Bytes()
}

//Bytes writes bytes
func (b *Writer) Bytes(v []byte) error {
	if err := b.Alloc(uint32(len(v))); err != nil {
		return err
	}
	_, err := b.Write(v)
	return err
}

//Float64 writes float64
func (b *Writer) Float64(v float64) error {
	bits := math.Float64bits(v)
	b.coder.PutUint64(b.buf, bits)
	_, err := b.Write(b.buf[:8])
	return err
}

//Float64s writes float64
func (b *Writer) Float64s(v []float64) error {
	if err := b.Alloc(uint32(len(v))); err != nil {
		return err
	}
	for _, i := range v {
		b.coder.PutUint64(b.buf, uint64(i))
		bits := math.Float64bits(i)
		b.coder.PutUint64(b.buf, bits)
		if _, err := b.Write(b.buf[:8]); err != nil {
			return err
		}
	}
	return nil
}

//NewWriter creates a binary writer
func NewWriter(coder binary.ByteOrder) *Writer {
	return &Writer{
		coder: coder,
		buf:   make([]byte, 8),
	}
}
