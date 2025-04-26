package bintly

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/viant/bintly/conv"
)

type (
	//Reader represents binary readers
	Reader struct {
		decAlloc decInt32s
		decMAlloc decUint16s
		decInts
		decUints
		decInt64s
		decUint64s
		decInt32s
		decUint32s
		decInt16s
		decUint16s
		decInt8s
		decUint8s
		decFloat64s
		decFloat32s
	}

	decInts     []int
	decUints    []uint
	decInt64s   []int64
	decUint64s  []uint64
	decInt32s   []int32
	decUint32s  []uint32
	decInt16s   []int16
	decUint16s  []uint16
	decInt8s    []int8
	decUint8s   []uint8
	decFloat64s []float64
	decFloat32s []float32
)

//Any returns value into source pointer, source has to be a pointer
func (r *Reader) Any(v interface{}) error {
	switch actual := v.(type) {
	case *int:
		r.Int(actual)
	case **int:
		r.IntPtr(actual)
	case *[]int:
		r.Ints(actual)
	case *uint:
		r.Uint(actual)
	case **uint:
		r.UintPtr(actual)
	case *[]uint:
		r.Uints(actual)
	case *int64:
		r.Int64(actual)
	case **int64:
		r.Int64Ptr(actual)
	case *[]int64:
		r.Int64s(actual)
	case *uint64:
		r.Uint64(actual)
	case **uint64:
		r.Uint64Ptr(actual)
	case *[]uint64:
		r.Uint64s(actual)
	case *int32:
		r.Int32(actual)
	case **int32:
		r.Int32Ptr(actual)
	case *[]int32:
		r.Int32s(actual)
	case *uint32:
		r.Uint32(actual)
	case **uint32:
		r.Uint32Ptr(actual)
	case *[]uint32:
		r.Uint32s(actual)
	case *uint16:
		r.Uint16(actual)
	case **uint16:
		r.Uint16Ptr(actual)
	case *[]uint16:
		r.Uint16s(actual)
	case *int16:
		r.Int16(actual)
	case **int16:
		r.Int16Ptr(actual)
	case *[]int16:
		r.Int16s(actual)
	case *uint8:
		r.Uint8(actual)
	case **uint8:
		r.Uint8Ptr(actual)
	case *[]uint8:
		r.Uint8s(actual)
	case *int8:
		r.Int8(actual)
	case **int8:
		r.Int8Ptr(actual)
	case *[]int8:
		r.Int8s(actual)
	case *float64:
		r.Float64(actual)
	case **float64:
		r.Float64Ptr(actual)
	case *[]float64:
		r.Float64s(actual)
	case *float32:
		r.Float32(actual)
	case **float32:
		r.Float32Ptr(actual)
	case *[]float32:
		r.Float32s(actual)
	case *bool:
		r.Bool(actual)
	case **bool:
		r.BoolPtr(actual)
	case *[]bool:
		r.Bools(actual)
	case *string:
		r.String(actual)
	case **string:
		r.StringPtr(actual)
	case *[]string:
		r.Strings(actual)
	case *time.Time:
		r.Time(actual)
	case **time.Time:
		r.TimePtr(actual)
	default:
		coder, ok := v.(Decoder)
		if ok {
			return r.Coder(coder)
		}
		return r.anyReflect(v)
	}
	return nil
}

//Alloc shifts allocation size (for repeated or pointers(nil:0,1))
func (r *Reader) Alloc() int32 {
	alloc := r.decAlloc[0]
	r.decAlloc = r.decAlloc[1:]
	return alloc
}

//MAlloc shifts allocation size (for repeated or pointers(nil:0,1))
func (r *Reader) MAlloc() uint16 {
	alloc := r.decMAlloc[0]
	r.decMAlloc = r.decMAlloc[1:]
	return alloc
}

//Int reads into *int
func (r *Reader) Int(v *int) {
	*v = r.decInts[0]
	r.decInts = r.decInts[1:]
}

//IntPtr reads into **int
func (r *Reader) IntPtr(v **int) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decInts[0]
	r.decInts = r.decInts[1:]
	*v = &i
}

//Ints reads into *[]int
func (r *Reader) Ints(vs *[]int) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decInts[:size]
	*vs = s
	r.decInts = r.decInts[size:]
}


//MInts reads medium size slice into *[]int
func (r *Reader) MInts(vs *[]int) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decInts[:size]
	*vs = s
	r.decInts = r.decInts[size:]
}

//Uint reads into *uint
func (r *Reader) Uint(v *uint) {
	*v = r.decUints[0]
	r.decUints = r.decUints[1:]
}

//UintPtr reads into **uint
func (r *Reader) UintPtr(v **uint) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUints[0]
	r.decUints = r.decUints[1:]
	*v = &i
}

//Uints reads into *[]uint
func (r *Reader) Uints(vs *[]uint) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decUints[:size]
	*vs = s
	r.decUints = r.decUints[size:]
}

//MUints reads into *[]uint
func (r *Reader) MUints(vs *[]uint) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decUints[:size]
	*vs = s
	r.decUints = r.decUints[size:]
}

//Int64 reads into *int64
func (r *Reader) Int64(v *int64) {
	*v = r.decInt64s[0]
	r.decInt64s = r.decInt64s[1:]
}

//Int64Ptr reads into **int64
func (r *Reader) Int64Ptr(v **int64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decInt64s[0]
	r.decInt64s = r.decInt64s[1:]
	*v = &i
}

//Int64s reads into *[]int64
func (r *Reader) Int64s(vs *[]int64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decInt64s[:size]
	*vs = s
	r.decInt64s = r.decInt64s[size:]
}

//MInt64s reads into *[]int64
func (r *Reader) MInt64s(vs *[]int64) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decInt64s[:size]
	*vs = s
	r.decInt64s = r.decInt64s[size:]
}

//Uint64 reads into *uint64
func (r *Reader) Uint64(v *uint64) {
	*v = r.decUint64s[0]
	r.decUint64s = r.decUint64s[1:]
}

//Uint64Ptr reads into **uint64
func (r *Reader) Uint64Ptr(v **uint64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUint64s[0]
	r.decUint64s = r.decUint64s[1:]
	*v = &i
}

//Uint64s reads into *[]uint64
func (r *Reader) Uint64s(vs *[]uint64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decUint64s[:size]
	*vs = s
	r.decUint64s = r.decUint64s[size:]
}


//MUint64s reads into *[]uint64
func (r *Reader) MUint64s(vs *[]uint64) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decUint64s[:size]
	*vs = s
	r.decUint64s = r.decUint64s[size:]
}

//Int32 reads into *int32
func (r *Reader) Int32(v *int32) {
	*v = r.decInt32s[0]
	r.decInt32s = r.decInt32s[1:]
}

//Int32Ptr reads into **int32
func (r *Reader) Int32Ptr(v **int32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decInt32s[0]
	r.decInt32s = r.decInt32s[1:]
	*v = &i
}

//Int32s reads into *[]int32
func (r *Reader) Int32s(vs *[]int32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decInt32s[:size]
	*vs = s
	r.decInt32s = r.decInt32s[size:]
}

//MInt32s reads into *[]int32
func (r *Reader) MInt32s(vs *[]int32) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decInt32s[:size]
	*vs = s
	r.decInt32s = r.decInt32s[size:]
}


//Uint32 reads into *uint32
func (r *Reader) Uint32(v *uint32) {
	*v = r.decUint32s[0]
	r.decUint32s = r.decUint32s[1:]
}

//Uint32Ptr reads into **uint32
func (r *Reader) Uint32Ptr(v **uint32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUint32s[0]
	r.decUint32s = r.decUint32s[1:]
	*v = &i
}

//Uint32s reads into  *[]uint32
func (r *Reader) Uint32s(vs *[]uint32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decUint32s[:size]
	*vs = s
	r.decUint32s = r.decUint32s[size:]
}


//Uint32s reads into  *[]uint32
func (r *Reader) MUint32s(vs *[]uint32) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decUint32s[:size]
	*vs = s
	r.decUint32s = r.decUint32s[size:]
}


//Int16 reads into *int16
func (r *Reader) Int16(v *int16) {
	*v = r.decInt16s[0]
	r.decInt16s = r.decInt16s[1:]
}

//Int16Ptr reads into **int16
func (r *Reader) Int16Ptr(v **int16) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decInt16s[0]
	r.decInt16s = r.decInt16s[1:]
	*v = &i
}

//Int16s reads into *[]int16
func (r *Reader) Int16s(vs *[]int16) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decInt16s[:size]
	*vs = s
	r.decInt16s = r.decInt16s[size:]
}

//Uint16 reads into *uint16
func (r *Reader) Uint16(v *uint16) {
	*v = r.decUint16s[0]
	r.decUint16s = r.decUint16s[1:]
}

//Uint16Ptr reads into **uint16
func (r *Reader) Uint16Ptr(v **uint16) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUint16s[0]
	r.decUint16s = r.decUint16s[1:]
	*v = &i
}

//Uint16s read into *[]uint16
func (r *Reader) Uint16s(vs *[]uint16) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decUint16s[:size]
	*vs = s
	r.decUint16s = r.decUint16s[size:]
}

//Int8 reads into *int8
func (r *Reader) Int8(v *int8) {
	*v = r.decInt8s[0]
	r.decInt8s = r.decInt8s[1:]
}

//Int8Ptr reads into **int8
func (r *Reader) Int8Ptr(v **int8) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decInt8s[0]
	r.decInt8s = r.decInt8s[1:]
	*v = &i
}

//Int8s reads into *[]int8
func (r *Reader) Int8s(vs *[]int8) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decInt8s[:size]
	*vs = s
	r.decInt8s = r.decInt8s[size:]
}

//Uint8 reads into *uint8
func (r *Reader) Uint8(v *uint8) {
	*v = r.decUint8s[0]
	r.decUint8s = r.decUint8s[1:]
}

//Uint8Ptr reads into **uint8
func (r *Reader) Uint8Ptr(v **uint8) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUint8s[0]
	r.decUint8s = r.decUint8s[1:]
	*v = &i
}

//Uint8s reads into *[]uint8
func (r *Reader) Uint8s(vs *[]uint8) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decUint8s[:size]
	*vs = s
	r.decUint8s = r.decUint8s[size:]
}

//MUint8s reads medium size slice (64k) into *[]uint8
func (r *Reader) MUint8s(vs *[]uint8) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	s := r.decUint8s[:size]
	*vs = s
	r.decUint8s = r.decUint8s[size:]
}


//Float64 reads into *float64
func (r *Reader) Float64(v *float64) {
	*v = r.decFloat64s[0]
	r.decFloat64s = r.decFloat64s[1:]
}

//Float64Ptr reads into **float64
func (r *Reader) Float64Ptr(v **float64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decFloat64s[0]
	r.decFloat64s = r.decFloat64s[1:]
	*v = &i
}

//Float64s reads into *[]float64
func (r *Reader) Float64s(vs *[]float64) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decFloat64s[:size]
	*vs = s
	r.decFloat64s = r.decFloat64s[size:]
}

//Float32 reads into *float32
func (r *Reader) Float32(v *float32) {
	*v = r.decFloat32s[0]
	r.decFloat32s = r.decFloat32s[1:]
}

//Float32Ptr reads into **float32
func (r *Reader) Float32Ptr(v **float32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decFloat32s[0]
	r.decFloat32s = r.decFloat32s[1:]
	*v = &i
}

//Float32s reads into *[]float32
func (r *Reader) Float32s(vs *[]float32) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	s := r.decFloat32s[:size]
	*vs = s
	r.decFloat32s = r.decFloat32s[size:]
}

//Bool reads into *bool
func (r *Reader) Bool(v *bool) {
	val := false
	i := r.decUint8s[0]
	r.decUint8s = r.decUint8s[1:]
	if i == 1 {
		val = true
	}
	*v = val
}

//BoolPtr reads into **bool
func (r *Reader) BoolPtr(v **bool) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	i := r.decUint8s[0]
	val := false
	r.decUint8s = r.decUint8s[1:]
	if i == 1 {
		val = true
	}
	*v = &val
}

//Bools reads into *[]bool
func (r *Reader) Bools(vs *[]bool) {
	size := int(r.Alloc())
	if size == 0 {
		return
	}
	var bools = make([]bool, size)
	for i := 0; i < size; i++ {
		bools[i] = r.decUint8s[i] == 1
	}
	r.decUint8s = r.decUint8s[size:]
	*vs = bools
}

//String reads into *string
func (r *Reader) String(v *string) {
	var bs []byte
	r.Uint8s(&bs)
	s := unsafeGetString(bs)
	*v = s
}

//StringPtr reads into **string
func (r *Reader) StringPtr(v **string) {
	var bs []byte
	r.Uint8s(&bs)
	s := unsafeGetString(bs)
	*v = &s
}

//Strings reads into *[]string
func (r *Reader) Strings(v *[]string) {
	size := r.Alloc()
	if size == 0 {
		return
	}
	var strings = make([]string, size)
	for i := 0; i < int(size); i++ {
		r.String(&strings[i])
	}
	*v = strings
}


//MString reads into *string
func (r *Reader) MString(v *string) {
	var bs []byte
	r.MUint8s(&bs)
	s := unsafeGetString(bs)
	*v = s
}

//MStringPtr reads into **string
func (r *Reader) MStringPtr(v **string) {
	var bs []byte
	r.MUint8s(&bs)
	s := unsafeGetString(bs)
	*v = &s
}

//MStrings reads into *[]string
func (r *Reader) MStrings(v *[]string) {
	size := r.MAlloc()
	if size == 0 {
		return
	}
	var strings = make([]string, size)
	for i := 0; i < int(size); i++ {
		r.MString(&strings[i])
	}
	*v = strings
}


//Time reads into *time.Time
func (r *Reader) Time(v *time.Time) {
	n := int64(0)
	r.Int64(&n)
	*v = time.Unix(0, n)
}

//TimePtr reads into **time.Time
func (r *Reader) TimePtr(v **time.Time) {
	var ptr *int64
	r.Int64Ptr(&ptr)
	if ptr == nil {
		return
	}
	t := time.Unix(0, *ptr)
	*v = &t
}

//Coder decodes coder
func (r *Reader) Coder(coder Decoder) error {
	size := r.Alloc()
	if allocator, ok := coder.(Allocator); ok {
		allocator.SetAlloc(size)
	}
	switch size {
	case NilSize, 0:
		return nil
	case 1:
		return coder.DecodeBinary(r)
	}
	for i := 0; i < int(size); i++ {
		if err := coder.DecodeBinary(r); err != nil {
			return err
		}
	}
	return nil
}

//FromBytes loads stream from bytes
func (r *Reader) FromBytes(data []byte) error {
	offset := 0
	offset = r.decAlloc.load(data, offset, codecAlloc)
	offset = r.decMAlloc.load(data, offset, codecMAlloc)
	offset = r.decInts.load(data, offset)
	offset = r.decFloat64s.load(data, offset)
	offset = r.decUint8s.load(data, offset)
	if data[offset] == codecEOF {
		return nil
	}
	offset = r.decInt32s.load(data, offset, codecInt32s)
	offset = r.decUint32s.load(data, offset)
	if data[offset] == codecEOF {
		return nil
	}
	offset = r.decFloat32s.load(data, offset)
	if data[offset] == codecEOF {
		return nil
	}
	offset = r.decUints.load(data, offset)
	offset = r.decInt64s.load(data, offset)
	offset = r.decUint64s.load(data, offset)
	offset = r.decInt16s.load(data, offset)
	offset = r.decUint16s.load(data, offset, codecUint16s)
	offset = r.decInt8s.load(data, offset)
	if data[offset] != codecEOF {
		return fmt.Errorf("corrupted bintly stream expected: %v, but had %v", codecEOF, data[offset])
	}
	return nil
}

func (r *Reader) anyReflect(v interface{}) error {
	rawType := reflect.TypeOf(v)
	if rawType.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, but had:%T", v)
	}
	rawType = rawType.Elem()
	value := reflect.ValueOf(v)
	isPointer := rawType.Kind() == reflect.Ptr
	if isPointer {
		rawType = rawType.Elem()
	}
	switch rawType.Kind() {
	case reflect.Struct:
		coder := structCoders.Get()
		defer structCoders.Put(coder)
		if err := coder.set(value, rawType); err != nil {
			return err
		}
		return r.Coder(coder)
	case reflect.Map:
		coder := mapCoders.Get()
		defer mapCoders.Put(coder)
		coder.set(value, rawType)
		return r.Coder(coder)
	case reflect.Slice:
		coder := sliceCoders.Get()
		defer sliceCoders.Put(coder)
		coder.set(value, rawType)
		return r.Coder(coder)
	default:
		//handles natives type aliases
		if nativeType := conv.MatchNative(rawType); nativeType != nil {
			native := reflect.New(*nativeType)
			if err := r.Any(native.Interface()); err != nil {
				return err
			}
			alias := native.Elem().Convert(rawType)
			if isPointer { //**T case
				actual := reflect.New(rawType)
				actual.Elem().Set(alias)
				value.Elem().Set(actual)
			} else {
				value.Elem().Set(alias)
			}
			return nil
		}
	}
	return fmt.Errorf("unsupproted readers type: %T", v)
}

func (s *decInts) load(data []byte, offset int) int {
	if data[offset] != codecInts {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Uint32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]int, size)
	GetInts(data[offset:], *s)
	offset += sizeIntInBytes * size
	return offset
}

func (s *decUints) load(data []byte, offset int) int {
	if data[offset] != codecUints {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Uint32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]uint, size)
	GetUints(data[offset:], *s)
	offset += sizeIntInBytes * size
	return offset
}

func (s *decInt64s) load(data []byte, offset int) int {
	if data[offset] != codecInt64s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]int64, size)
	GetInt64s(data[offset:], *s)
	offset += size64bitsInBytes * size
	return offset
}

func (s *decUint64s) load(data []byte, offset int) int {
	if data[offset] != codecUint64s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Uint32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]uint64, size)
	GetUint64s(data[offset:], *s)
	offset += size64bitsInBytes * size
	return offset
}

func (s *decInt32s) load(data []byte, offset int, codec uint8) int {
	if data[offset] != codec {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Uint32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]int32, size)
	GetInt32s(data[offset:], *s)
	offset += size32bitsInBytes * size
	return offset
}

func (s *decUint32s) load(data []byte, offset int) int {
	if data[offset] != codecUint32s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Uint32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]uint32, size)
	GetUint32s(data[offset:], *s)
	offset += size32bitsInBytes * size
	return offset
}

func (s *decInt16s) load(data []byte, offset int) int {
	if data[offset] != codecInt16s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]int16, size)
	GetInt16s(data[offset:], *s)
	offset += size16bitsInBytes * size
	return offset
}

func (s *decUint16s) load(data []byte, offset int, codec uint8) int {
	if data[offset] != codec {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]uint16, size)
	GetUint16s(data[offset:], *s)
	offset += size16bitsInBytes * size
	return offset
}

func (s *decInt8s) load(data []byte, offset int) int {
	if data[offset] != codecInt8s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]int8, size)
	GetInt8s(data[offset:], *s)
	offset += size8bitsInBytes * size
	return offset
}

func (s *decUint8s) load(data []byte, offset int) int {
	if data[offset] != codecUint8s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]uint8, size)
	GetUint8s(data[offset:], *s)
	offset += size8bitsInBytes * size
	return offset
}

func (s *decFloat64s) load(data []byte, offset int) int {
	if data[offset] != codecFloat64s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]float64, size)
	GetFloat64s(data[offset:], *s)
	offset += size64bitsInBytes * size
	return offset
}

func (s *decFloat32s) load(data []byte, offset int) int {
	if data[offset] != codecFloat32s {
		return offset
	}
	offset += size8bitsInBytes
	size := int(Int32(data[offset:]))
	offset += size32bitsInBytes
	*s = make([]float32, size)
	GetFloat32s(data[offset:], *s)
	offset += size32bitsInBytes * size
	return offset
}

//Readers represents readers pool
type Readers struct {
	sync.Pool
}

//Get returns a readers
func (p *Readers) Get() *Reader {
	codec := p.Pool.Get()
	return codec.(*Reader)
}

//NewReaders creates a readers pool
func NewReaders() *Readers {
	return &Readers{
		Pool: sync.Pool{
			New: func() interface{} {
				return &Reader{}
			},
		},
	}
}
