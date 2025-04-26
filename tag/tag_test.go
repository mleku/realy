package tag

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/chk"
	"realy.lol/log"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b, bool, bc []byte
	for _ = range 100 {
		n := frand.Intn(8)
		tg := NewWithCap(n)
		for _ = range n {
			b1 := make([]byte, frand.Intn(8))
			_, _ = frand.Read(b1)
			tg.field = append(tg.field, b1)
		}
		// log.I.S(tg)
		b = tg.Marshal(b)
		bool = make([]byte, len(b))
		copy(bool, b)
		tg2 := NewWithCap(n)
		rem, err := tg2.Unmarshal(b)
		// log.I.S(tg2)
		if chk.E(err) {
			t.Fatal(err)
		}
		bc = tg2.Marshal(bc)
		// log.I.ToSliceOfBytes("\n\norig\n%s\n\ncopy\n%s\n", bo, bc)
		if !bytes.Equal(bool, bc) {
			t.Fatalf("got\n%s\nwant\n%s", bool, bc)
		}
		if len(rem) != 0 {
			t.Fatalf("len(rem)!=0:\n%s", rem)
		}
		if !tg.Equal(tg2) {
			t.Fatalf("got\n%s\nwant\n%s", tg2, tg)
		}
		b, bool, bc = b[:0], bool[:0], bc[:0]
	}
}

func TestMarshalUnmarshalZeroLengthTag(t *testing.T) {
	empty := []byte("[\"a\"]")
	var b []byte
	tg := &T{}
	b, _ = tg.Unmarshal(empty)
	b = tg.Marshal(b)
	if !bytes.Equal(empty, b) {
		t.Fatalf("got\n%s\nwant\n%s", b, empty)
	}
	empty = []byte("[]")
	tg = &T{}
	b, _ = tg.Unmarshal(empty)
	b = tg.Marshal(b)
	if !bytes.Equal(empty, b) {
		t.Fatalf("got\n%s\nwant\n%s", b, empty)
	}
}

func BenchmarkMarshalUnmarshal(bb *testing.B) {
	b := make([]byte, 0, 40000000)
	n := 4096
	tg := NewWithCap(n)
	for _ = range n {
		b1 := make([]byte, 128)
		_, _ = frand.Read(b1)
		tg.field = append(tg.field, b1)
	}
	bb.Run("tag.Marshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			b = tg.Marshal(b)
			b = b[:0]
			tg.Clear()
		}
	})
	bb.Run("tag.MarshalUnmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		var tg2 T
		for i := 0; i < bb.N; i++ {
			b = tg.Marshal(b)
			_, _ = tg2.Unmarshal(b)
			b = b[:0]
			tg.Clear()
		}
	})
}

func TestT_Clone_Equal(t *testing.T) {
	for _ = range 100 {
		n := frand.Intn(64) + 2
		t1 := NewWithCap(n)
		for _ = range n {
			f := make([]byte, frand.Intn(128)+2)
			_, _ = frand.Read(f)
			t1.field = append(t1.field, f)
		}
		t2 := t1.Clone()
		if !t1.Equal(t2) {
			log.E.S(t1, t2)
			t.Fatal("not equal")
		}
	}
}
