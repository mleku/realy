package tags

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/hex"
	"realy.lol/tag"
)

func TestMarshalUnmarshal(t *testing.T) {
	var b, rem []byte
	var err error
	for _ = range 100000 {
		n := frand.Intn(3)
		tgs := &T{}
		for _ = range n {
			n1 := frand.Intn(5)
			tg := tag.NewWithCap(n)
			for _ = range n1 {
				b1 := make([]byte, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg = tg.Append(b1)
			}
			tgs.element = append(tgs.element, tg)
		}
		b = tgs.Marshal(b)
		bb := make([]byte, len(b))
		copy(bb, b)
		ta := &T{}
		rem, err = ta.Unmarshal(b)
		if chk.E(err) {
			t.Fatal(err)
		}
		var bc []byte
		bc = ta.Marshal(bc)
		if !bytes.Equal(bb, bc) {
			t.Fatalf("got\n%s\nwant\n%s\n", bc, bb)
		}
		b, rem, bb, bc = b[:0], rem[:0], bb[:0], bc[:0]
	}
}

func TestEmpty(t *testing.T) {
	var err error
	var b0, bc, b1 []byte
	var empty, empty1, empty2 *T
	_, _, _ = empty, empty1, empty2
	empty = New()
	b0 = empty.Marshal(b0)
	bc = make([]byte, len(b0))
	copy(bc, b0)
	empty1 = New()
	if b0, err = empty1.Unmarshal(b0); chk.E(err) {
		t.Fatal(err)
	}
	empty2 = New()
	b1 = empty2.Marshal(b1)
	if !bytes.Equal(bc, b1) {
		t.Fatalf("'%s' == '%s' -> %v", bc, b1, bytes.Equal(bc, b1))
	}
	b0, bc, b1 = b0[:0], bc[:0], b1[:0]
	empty = New(&tag.T{})
	b0 = empty.Marshal(b0)
	bc = make([]byte, len(b0))
	copy(bc, b0)
	empty1 = New()
	if b0, err = empty1.Unmarshal(b0); chk.E(err) {
		t.Fatal(err)
	}
	empty2 = New()
	b1 = empty1.Marshal(b1)
	if !bytes.Equal(bc, b1) {
		t.Fatalf("'%s' == '%s' -> %v", bc, b1, bytes.Equal(bc, b1))
	}
	b0, bc, b1 = b0[:0], bc[:0], b1[:0]
}

func BenchmarkMarshalUnmarshal(bb *testing.B) {
	var b, rem []byte
	var err error
	bb.Run("tag.Marshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make([]byte, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg = tg.Append(b1)
					// tg.Field = append(tg.Field, b1)
				}
				tgs.element = append(tgs.element, tg)
			}
			b = tgs.Marshal(b)
			b = b[:0]
		}
	})
	bb.Run("tag.MarshalUnmarshal", func(bb *testing.B) {
		bb.ReportAllocs()
		for i := 0; i < bb.N; i++ {
			n := frand.Intn(40) + 2
			tgs := New()
			for _ = range n {
				n1 := frand.Intn(40) + 2
				tg := tag.NewWithCap(n)
				for _ = range n1 {
					b1 := make([]byte, frand.Intn(40)+2)
					_, _ = frand.Read(b1)
					tg = tg.Append(b1)
				}
				tgs.element = append(tgs.element, tg)
			}
			b = tgs.Marshal(b)
			ta := New()
			rem, err = ta.Unmarshal(b)
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
				b1 := make([]byte, frand.Intn(40)+2)
				_, _ = frand.Read(b1)
				tg = tg.Append(b1)
			}
			t1.element = append(t1.element, tg)
		}
		t2 := t1.Clone()
		if !t1.Equal(t2) {
			log.E.S(t1, t2)
			t.Fatal("not equal")
		}
	}
}

func TestTagHelpers(t *testing.T) {
	tags := New(
		tag.New("x"),
		tag.New("p", "abcdef", "wss://x.com"),
		tag.New("p", "123456", "wss://y.com"),
		tag.New("e", "eeeeee"),
		tag.New("e", "ffffff"),
	)

	if tags.GetFirst(tag.New("x")) == nil {
		t.Error("failed to get existing prefix")
	}
	if tags.GetFirst(tag.New("x", "")) != nil {
		t.Error("got with wrong prefix")
	}
	if tags.GetFirst(tag.New("p", "abcdef", "wss://")) == nil {
		t.Error("failed to get with existing prefix")
	}
	if tags.GetFirst(tag.New("p", "abcdef", "")) == nil {
		t.Error("failed to get with existing prefix (blank last string)")
	}
	if string(tags.GetLast(tag.New("e")).S(1)) != "ffffff" {
		t.Error("failed to get last")
	}
	if tags.GetAll(tag.New("e", "")).Len() != 2 {
		t.Error("failed to get all")
	}
	if tags.AppendUnique(tag.New("e", "ffffff")).Len() != 5 {
		t.Error("append unique changed the array size when existed")
	}
	if tags.AppendUnique(tag.New("e", "bbbbbb")).Len() != 6 {
		t.Error("append unique failed to append when didn't exist")
	}
	if string(tags.AppendUnique(tag.New("e", "eeeeee")).GetTagElement(4).S(1)) != "ffffff" {
		t.Error("append unique changed the order")
	}
	if string(tags.AppendUnique(tag.New("e", "eeeeee")).GetTagElement(3).S(1)) != "eeeeee" {
		t.Error("append unique changed the order")
	}
}

func TestT_ContainsAny(t *testing.T) {
	var v, a, b, c []byte
	var err error
	v, err = hex.Dec("4c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f")
	a, err = hex.Dec("3c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f")
	b, err = hex.Dec("2c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f")
	c, err = hex.Dec("1c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f")
	w := tag.New([]byte{'b'}, v, a, b, c)
	x := tag.New([]byte{'b'}, c, b, a)
	y := tag.New([]byte{'b'}, b, a, c)
	z := tag.New([]byte{'b'}, v)
	tt := New(w, x, y)
	ttt := New(x, y)
	_, _, _, _, _ = v, err, z, tt, ttt
	// log.I.S(tt.ContainsAny(by{'b'}, z))
	// log.I.S(ttt.ContainsAny(by{'b'}, z))

}
