package tags

import (
	"testing"

	"lukechampine.com/frand"
	"mleku.dev/tag"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b, rem B
	var err error
	for _ = range 100000 {
		n := frand.Intn(3)
		tgs := &T{}
		for _ = range n {
			n1 := frand.Intn(5)
			tg := tag.NewWithCap(n)
			for _ = range n1 {
				b1 := make(B, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg.Field = append(tg.Field, b1)
			}
			tgs.T = append(tgs.T, tg)
		}
		b, _ = tgs.MarshalJSON(b)
		bo := make(B, len(b))
		copy(bo, b)
		ta := &T{}
		rem, err = ta.UnmarshalJSON(b)
		if chk.E(err) {
			t.Fatal(err)
		}
		var bc B
		bc, _ = ta.MarshalJSON(bc)
		if !equals(bo, bc) {
			t.Fatalf("got\n%s\nwant\n%s\n", bc, bo)
		}
		b, rem, bo, bc = b[:0], rem[:0], bo[:0], bc[:0]
	}
}

func TestEmpty(t *testing.T) {
	var err error
	var b0, bc, b1 B
	var empty, empty1, empty2 *T
	_, _, _ = empty, empty1, empty2
	empty = New()
	if b0, err = empty.MarshalJSON(b0); chk.E(err) {
		t.Fatal(err)
	}
	bc = make(B, len(b0))
	copy(bc, b0)
	empty1 = New()
	if b0, err = empty1.UnmarshalJSON(b0); chk.E(err) {
		t.Fatal(err)
	}
	empty2 = New()
	if b1, err = empty2.MarshalJSON(b1); chk.E(err) {
		t.Fatal(err)
	}
	if !equals(bc, b1) {
		t.Fatalf("'%s' == '%s' -> %v", bc, b1, equals(bc, b1))
	}
	b0, bc, b1 = b0[:0], bc[:0], b1[:0]
	empty = New(&tag.T{})
	if b0, err = empty.MarshalJSON(b0); chk.E(err) {
		t.Fatal(err)
	}
	bc = make(B, len(b0))
	copy(bc, b0)
	empty1 = New()
	if b0, err = empty1.UnmarshalJSON(b0); chk.E(err) {
		t.Fatal(err)
	}
	empty2 = New()
	if b1, err = empty1.MarshalJSON(b1); chk.E(err) {
		t.Fatal(err)
	}
	if !equals(bc, b1) {
		t.Fatalf("'%s' == '%s' -> %v", bc, b1, equals(bc, b1))
	}
	b0, bc, b1 = b0[:0], bc[:0], b1[:0]
}

func BenchmarkMarshalJSONUnmarshalJSON(bb *testing.B) {
	var b, rem B
	var err error
	bb.Run("tag.MarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg.Field = append(tg.Field, b1)
				}
				tgs.T = append(tgs.T, tg)
			}
			b, _ = tgs.MarshalJSON(b)
			b = b[:0]
		}
	})
	bb.Run("tag.MarshalJSONUnmarshalJSON", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make(B, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg.Field = append(tg.Field, b1)
				}
				tgs.T = append(tgs.T, tg)
			}
			b, _ = tgs.MarshalJSON(b)
			ta := New()
			rem, err = ta.UnmarshalJSON(b)
			if chk.E(err) {
				bb.Fatal(err)
			}
			if len(rem) != 0 {
				bb.Fatalf("len(rem)!=0:\n%s", rem)
			}
			if !tgs.Equal(ta) {
				bb.Fatalf("got\n%v\nwant\n%v", ta, tgs)
			}
			b = b[:0]
		}
	})
}
func TestT_Clone_Equal(t *testing.T) {
	for _ = range 10 {
		n := frand.Intn(40) + 2
		t1 := New()
		for _ = range n {
			n1 := frand.Intn(40) + 2
			tg := tag.NewWithCap(n)
			for _ = range n1 {
				b1 := make(B, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg.Field = append(tg.Field, b1)
			}
			t1.T = append(t1.T, tg)
		}
		t2 := t1.Clone()
		if !t1.Equal(t2) {
			log.E.S(t1, t2)
			t.Fatal("not equal")
		}
	}
}
