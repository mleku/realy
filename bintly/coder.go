package bintly

import (
	"reflect"
	"sync"

	"github.com/viant/bintly/conv"
)

type structFields struct {
	//indexes of the exported fields
	indexes     []int
	convertible []convField
}

type convField struct {
	index  int
	origin reflect.Type
	native reflect.Type
}

//structCoder represents a struct coder
type structCoder struct {
	mapped map[reflect.Type]*structFields
	ptr    *reflect.Value
	isNil  bool
	v      reflect.Value
	t      reflect.Type
	fields *structFields
}

func (c *structCoder) Alloc() int32 {
	if c.isNil {
		return NilSize
	}
	return 1
}

//SetAlloc set allocation, if zero the pointer to struct is nil
func (c *structCoder) SetAlloc(allocation int32) {
	if allocation <= 0 {
		c.ptr = nil
		return
	}
	if c.v.Kind() == reflect.Ptr {
		v := reflect.New(c.t)
		c.v.Set(v)
		c.v = v.Elem()
	}
}

func (c *structCoder) setFields(t reflect.Type) error {
	if fields, ok := c.mapped[t]; ok {
		c.fields = fields
		return nil
	}
	fields := &structFields{
		indexes:     make([]int, 0),
		convertible: make([]convField, 0),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if !conv.IsNative(f.Type) {
			if native := conv.MatchNative(f.Type); native != nil {
				fields.convertible = append(fields.convertible, convField{
					index:  i,
					origin: f.Type,
					native: *native,
				})
				continue
			}
		}
		fields.indexes = append(fields.indexes, i)
	}
	c.mapped[t] = fields
	c.fields = c.mapped[t]
	return nil
}

func (c *structCoder) set(v reflect.Value, t reflect.Type) error {
	c.isNil = false
	c.t = t
	c.v = v
	c.ptr = nil
	if v.Kind() == reflect.Ptr {
		c.ptr = &v
		c.isNil = v.IsNil()
		if !c.isNil {
			c.v = v.Elem()
		}
	}
	return c.setFields(t)
}

//EncodeBinary writes struct to stream
func (c *structCoder) EncodeBinary(stream *Writer) error {
	for _, i := range c.fields.indexes {
		v := c.v.Field(i).Interface()
		if err := stream.Any(v); err != nil {
			return err
		}
	}
	for _, f := range c.fields.convertible {
		v := c.v.Field(f.index).Convert(f.native)
		if err := stream.Any(v.Interface()); err != nil {
			return err
		}
	}
	return nil
}

//DecodeBinary decodes struct from reader
func (c *structCoder) DecodeBinary(stream *Reader) error {
	if c.ptr == nil {
		return nil
	}
	if c.ptr.IsNil() {
		c.v = reflect.New(c.t).Elem()
		c.ptr.Elem().Set(c.v)
	}

	for _, i := range c.fields.indexes {
		v := c.v.Field(i).Addr().Interface()
		if err := stream.Any(v); err != nil {
			return err
		}
	}
	for _, f := range c.fields.convertible {
		v := reflect.New(f.native)
		if err := stream.Any(v.Interface()); err != nil {
			return err
		}
		c.v.Field(f.index).Set(v.Elem().Convert(f.origin))
	}
	return nil
}

type structCoderPool struct {
	sync.Pool
}

func (s *structCoderPool) Get() *structCoder {
	return s.Pool.Get().(*structCoder)
}

func newStructCoderPool() *structCoderPool {
	return &structCoderPool{

		Pool: sync.Pool{
			New: func() interface{} {

				return &structCoder{
					mapped: make(map[reflect.Type]*structFields),
				}
			},
		},
	}
}

var structCoders = newStructCoderPool()

type sliceCoder struct {
	ptr      *reflect.Value
	index    int
	isNil    bool
	v        reflect.Value
	elemType reflect.Type //element type
}

func (c *sliceCoder) set(v reflect.Value, t reflect.Type) {
	c.index = 0
	c.isNil = v.IsNil()
	c.v = v
	if v.Kind() == reflect.Ptr {
		c.ptr = &v
		c.isNil = c.ptr.IsNil()
		if !c.isNil {
			c.v = v.Elem()
		}
	}
	c.elemType = t.Elem()
}

//Alloc returns slice size
func (c *sliceCoder) Alloc() int32 {
	if c.isNil {
		return NilSize
	}
	return int32(c.v.Len())
}

//SetAlloc set allocation, if zero the pointer to struct is nil
func (c *sliceCoder) SetAlloc(allocation int32) {
	if allocation < 0 {
		return
	}
	c.v = reflect.MakeSlice(reflect.SliceOf(c.elemType), int(allocation), int(allocation))
	c.ptr.Elem().Set(c.v)
}

//EncodeBinary writes slice to stream
func (c *sliceCoder) EncodeBinary(stream *Writer) error {
	item := c.v.Index(c.index)
	c.index++
	return stream.Any(item.Interface())
}

//DecodeBinary reads slice from stream
func (c *sliceCoder) DecodeBinary(stream *Reader) error {
	elem := reflect.New(c.elemType)
	if err := stream.Any(elem.Interface()); err != nil {
		return err
	}
	c.v.Index(c.index).Set(elem.Elem())
	c.index++
	return nil
}

type sliceCoderPool struct {
	sync.Pool
}

func (s *sliceCoderPool) Get() *sliceCoder {
	return s.Pool.Get().(*sliceCoder)
}

func newSliceCoderPool() *sliceCoderPool {
	return &sliceCoderPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return &sliceCoder{}
			},
		},
	}
}

var sliceCoders = newSliceCoderPool()

//mapCoder represents a map coder
type mapCoder struct {
	ptr  *reflect.Value
	v    reflect.Value
	t    reflect.Type
	k    reflect.Type
	e    reflect.Type
	iter *reflect.MapIter
}

func (c *mapCoder) set(v reflect.Value, t reflect.Type) {
	c.ptr = nil
	c.v = v
	if v.Kind() == reflect.Ptr {
		c.ptr = &v
	}
	c.t = t
	c.k = t.Key()
	c.e = t.Elem()
}

//Alloc returns slice size
func (c *mapCoder) Alloc() int32 {
	if c.v.IsNil() {
		return NilSize
	}
	c.iter = c.v.MapRange()
	return int32(c.v.Len())
}

//SetAlloc set allocation, if zero the pointer to struct is nil
func (c *mapCoder) SetAlloc(allocation int32) {
	if allocation < 0 {
		return
	}
	c.v = reflect.MakeMapWithSize(c.t, int(allocation))
	c.ptr.Elem().Set(c.v)

}

//EncodeBinary writes slice to stream
func (c *mapCoder) EncodeBinary(stream *Writer) error {
	if !c.iter.Next() {
		return nil
	}
	if err := stream.Any(c.iter.Key().Interface()); err != nil {
		return err
	}
	return stream.Any(c.iter.Value().Interface())
}

//DecodeBinary reads slice from stream
func (c *mapCoder) DecodeBinary(stream *Reader) error {
	key := reflect.New(c.k)
	if err := stream.Any(key.Interface()); err != nil {
		return err
	}
	val := reflect.New(c.e)
	if err := stream.Any(val.Interface()); err != nil {
		return err
	}
	c.v.SetMapIndex(key.Elem(), val.Elem())
	return nil
}

type mapCoderPool struct {
	sync.Pool
}

func (s *mapCoderPool) Get() *mapCoder {
	return s.Pool.Get().(*mapCoder)
}

func newMapCoderPool() *mapCoderPool {
	return &mapCoderPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return &mapCoder{}
			},
		},
	}
}

var mapCoders = newMapCoderPool()
