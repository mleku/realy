package stress

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"math"
	"testing"

	"github.com/fxamacker/cbor"
	"github.com/stretchr/testify/assert"
	"github.com/viant/bintly"
	"github.com/vmihailenco/msgpack"
)

var b1 = BenchStruct{
	A1: math.MaxInt32,
	A2: "this is benchmark test",
	A3: true,
	A4: 3.3333,
	A5: intSlice,
	A6: stringSlice,
	A7: float64Slice,
	A8: uint8Slice,
}

func Test_BenchStruct(t *testing.T) {

	{ // test custom binary
		data, err := b1.ToBytes()
		assert.Nil(t, err)
		clone := BenchStruct{}
		clone.FromBytes(data)
		assert.EqualValues(t, clone, b1)
	}
	{ // test custom bintly
		data, err := bintly.Marshal(&b1)
		assert.Nil(t, err)
		clone := BenchStruct{}
		err = bintly.Unmarshal(data, &clone)
		assert.EqualValues(t, clone, b1)
	}
	{ // test bintly reflect
		alias := BenchStructAlias(b1)
		data, err := bintly.Marshal(&alias)
		assert.Nil(t, err)
		clone := BenchStructAlias{}
		err = bintly.Unmarshal(data, &clone)
		assert.EqualValues(t, clone, alias)
	}
	{ // test gob
		var buf bytes.Buffer
		dec := gob.NewDecoder(&buf)
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(&b1)
		assert.Nil(t, err)
		clone := BenchStruct{}
		err = dec.Decode(&clone)
		assert.Nil(t, err)
		assert.EqualValues(t, b1, clone)
	}
	{ // test cobr reflect
		alias := BenchStructAlias(b1)
		data, err := cbor.Marshal(&alias, cbor.EncOptions{})
		assert.Nil(t, err)
		clone := BenchStructAlias{}
		err = cbor.Unmarshal(data, &clone)
		assert.EqualValues(t, clone, b1)
	}

}

func BenchmarkUnmarshalBintly(b *testing.B) {
	data, err := bintly.Marshal(&b1)
	assert.Nil(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		err = bintly.Unmarshal(data, &c1)
	}
	assert.Nil(b, err)
}

func BenchmarkMarshalBintly(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = bintly.Marshal(b1)
	}
}

func BenchmarkUnmarshalBintlyReflect(b *testing.B) {
	var a1 = BenchStructAlias(b1)
	data, err := bintly.Marshal(&a1)
	assert.Nil(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var b1 BenchStructAlias
		err = bintly.Unmarshal(data, &b1)

	}
	assert.Nil(b, err)
}

func BenchmarkMarshalBintlyReflect(b *testing.B) {
	var a1 = BenchStructAlias(b1)
	b.ResetTimer()
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = bintly.Marshal(&a1)
	}
	assert.Nil(b, err)
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	data, _ := b1.ToBytes()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		c1.FromBytes(data)
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = b1.ToBytes()
	}
	assert.Nil(b, err)
}

func BenchmarkUnMarshalGob(b *testing.B) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(b1)
	if err != nil {
		assert.NotNil(b, err)
	}
	data := buf.Bytes()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		dec := gob.NewDecoder(bytes.NewReader(data))
		err = dec.Decode(&c1)

	}
	assert.Nil(b, err, err)
}

func BenchmarkMarshalGob(b *testing.B) {
	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(b1)

	}
	assert.Nil(b, err)
}

func BenchmarkUnmarshalCbor(b *testing.B) {
	data, err := cbor.Marshal(b1, cbor.EncOptions{})
	if !assert.Nil(b, err) {
		return
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		err = cbor.Unmarshal(data, &c1)
	}
	assert.Nil(b, err)
}

func BenchmarkMarshalCbor(b *testing.B) {
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = cbor.Marshal(b1, cbor.EncOptions{})
	}
	assert.Nil(b, err)
}

func BenchmarkUnmarshalMsgPack(b *testing.B) {
	data, err := msgpack.Marshal(b1)
	if !assert.Nil(b, err) {
		return
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		err = msgpack.Unmarshal(data, &c1)
	}
	assert.Nil(b, err)
}

func BenchmarkMarshalMsgPack(b *testing.B) {
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = msgpack.Marshal(b1)
	}
	assert.Nil(b, err)
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	data, err := json.Marshal(b1)
	assert.Nil(b, err)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c1 := BenchStruct{}
		err = json.Unmarshal(data, &c1)
	}
	assert.Nil(b, err)

}
func BenchmarkJSONMarshal(b *testing.B) {
	b.ReportAllocs()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = json.Marshal(b1)
	}
	assert.Nil(b, err)
}
