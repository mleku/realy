package bintly

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructCoder_DecodeBinary(t *testing.T) {
	type Foo struct {
		ID   int
		Name string
	}
	type intAlias int
	ia := intAlias(3023)

	var (
		useCases = []struct {
			description string
			value       interface{}
		}{
			{
				description: "int/uint types",
				value: struct {
					I   int
					U   uint
					Is  []int
					Uis []uint
				}{1, 2, intSlice, uintSlice},
			},
			{
				description: "int64/uint64 types",
				value: struct {
					I int64
					U uint64
				}{1000, 3000},
			},
			{
				description: "int32/uint32 types",
				value: struct {
					I int32
					U uint32
				}{-1500, 30777},
			},
			{
				description: "int16/uint16 types",
				value: struct {
					I int16
					U uint16
				}{-1544, 664},
			},
			{
				description: "int16/uint16 types",
				value: struct {
					I int16
					U uint16
				}{-15, 255},
			},
			{
				description: "float64/float32 types",
				value: struct {
					F1 float64
					F2 float32
				}{0.1, 3.222},
			},
			{
				description: "bool type",
				value: struct {
					B    bool
					BPtr *bool
				}{true, &boolSlice[0]},
			},
			{
				description: "alias type",
				value: struct {
					I     intAlias
					Bytes []byte
					X     *intAlias
					Y     []intAlias
					Z     []*intAlias
				}{102, []byte("123"), &ia, []intAlias{1, 2, 3}, []*intAlias{&ia}},
			},
			{
				description: "anonymous type",
				value: struct {
					Foo
					Items []Foo
				}{Foo: Foo{1, "name"}, Items: []Foo{{10, "test"}}},
			},
			{
				description: "map[string]int type",
				value: struct {
					M map[string]int
					Z map[int]intAlias
				}{map[string]int{
					"k1": 1,
					"k2": 2,
				}, nil},
			},
			{
				description: "nil slice",
				value: struct {
					A []intAlias
					B []intAlias
				}{[]intAlias{1, 2}, nil},
			},
			{
				description: "nil struct",
				value: struct {
					A *Foo
					B *Foo
				}{A: &Foo{}, B: nil},
			},
		}
	)

	for _, useCase := range useCases {
		data, err := Marshal(useCase.value)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		actualPtr := reflect.New(reflect.TypeOf(useCase.value))
		err = Unmarshal(data, actualPtr.Interface())
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		assert.EqualValues(t, useCase.value, actualPtr.Elem().Interface(), useCase.description)

	}

}

func Test_CustomSlice(t *testing.T) {
	var aSlice = customSlice{"a", "b", "c"}
	data, err := Marshal(&aSlice)
	assert.Nil(t, err)
	var clone customSlice
	err = Unmarshal(data, &clone)
	assert.Nil(t, err)
	assert.EqualValues(t, clone, aSlice)
}

type customSlice []string

//DecodeBinary decodes data to binary stream
func (e *customSlice) DecodeBinary(stream *Reader) error {
	size := int(stream.Alloc())
	if size == NilSize {
		return nil
	}
	*e = make([]string, size)
	for i := 0; i < size; i++ {
		item := ""
		stream.String(&item)
		(*e)[i] = item
	}
	return nil
}

//EncodeBinary encodes data from binary stream
func (e *customSlice) EncodeBinary(stream *Writer) error {
	if *e == nil {
		stream.Alloc(NilSize)
		return nil
	}
	stream.Alloc(int32(len(*e)))
	for i := range *e {
		stream.String((*e)[i])
	}
	return nil
}


func Test_CustomMap(t *testing.T) {
	var aMap = customMap{"a": []uint32{1, 2}, "b": []uint32{2, 3}, "c": []uint32{3, 4}}
	data, err := Marshal(&aMap)
	assert.Nil(t, err)
	var clone customMap
	err = Unmarshal(data, &clone)
	assert.Nil(t, err)
	assert.EqualValues(t, clone, aMap)
}

type customMap map[string][]uint32

//DecodeBinary decodes data to binary stream
func (e *customMap) DecodeBinary(stream *Reader) error {
	size := int(stream.Alloc())
	if size == NilSize {
		return nil
	}
	*e = make(map[string][]uint32, size)
	for i := 0; i < size; i++ {
		var key string
		var val []uint32
		stream.MString(&key)
		stream.MUint32s(&val)
		(*e)[key] = val
	}
	return nil
}

//EncodeBinary encodes data from binary stream
func (e *customMap) EncodeBinary(stream *Writer) error {
	if *e == nil {
		stream.Alloc(NilSize)
		return nil
	}
	stream.Alloc(int32(len(*e)))
	for k, v := range *e {
		stream.MString(k)
		stream.MUint32s(v)
	}
	return nil
}
