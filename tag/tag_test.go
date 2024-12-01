package tag

import (
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	var b, bo, bc by
	for _ = range 100 {
		n := frand.Intn(8)
		tg := NewWithCap(n)
		for _ = range n {
			b1 := make(by, frand.Intn(8))
			_, _ = frand.Read(b1)
			tg.field = append(tg.field, b1)
		}
		// log.I.S(tg)
		b, _ = tg.MarshalJSON(b)
		bo = make(by, len(b))
		copy(bo, b)
		tg2 := NewWithCap(n)
		rem, err := tg2.UnmarshalJSON(b)
		// log.I.S(tg2)
		if chk.E(err) {
			t.Fatal(err)
		}
		bc, _ = tg2.MarshalJSON(bc)
		log.I.F("\n\norig\n%s\n\ncopy\n%s\n", bo, bc)
		if !equals(bo, bc) {
			t.Fatalf("got\n%s\nwant\n%s", bo, bc)
		}
		if len(rem) != 0 {
			t.Fatalf("len(rem)!=0:\n%s", rem)
		}
		if !tg.Equal(tg2) {
			t.Fatalf("got\n%s\nwant\n%s", tg2, tg)
		}
		b, bo, bc = b[:0], bo[:0], bc[:0]
	}
}

func TestMarshalUnmarshalZeroLengthTag(t *testing.T) {
	var err er
	empty := by("[\"a\"]")
	var b by
	tg := &T{}
	b, _ = tg.UnmarshalJSON(empty)
	if b, err = tg.MarshalJSON(b); chk.E(err) {
		t.Fatal(err)
	}
	if !equals(empty, b) {
		t.Fatalf("got\n%s\nwant\n%s", b, empty)
	}
	empty = by("[]")
	tg = &T{}
	b, _ = tg.UnmarshalJSON(empty)
	if b, err = tg.MarshalJSON(b); chk.E(err) {
		t.Fatal(err)
	}
	if !equals(empty, b) {
		t.Fatalf("got\n%s\nwant\n%s", b, empty)
	}
}

func BenchmarkMarshalJSONUnmarshalJSON(bb *testing.B) {
	b := make(by, 0, 40000000)
	n := 4096
	tg := NewWithCap(n)
	for _ = range n {
		b1 := make(by, 128)
		_, _ = frand.Read(b1)
		tg.field = append(tg.field, b1)
	}
	bb.Run("tag.MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			b, _ = tg.MarshalJSON(b)
			b = b[:0]
			tg.Clear()
		}
	})
	bb.Run("tag.MarshalJSONUnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		var tg2 T
		for i := 0; i < bb.N; i++ {
			b, _ = tg.MarshalJSON(b)
			_, _ = tg2.UnmarshalJSON(b)
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
			f := make(by, frand.Intn(128)+2)
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
