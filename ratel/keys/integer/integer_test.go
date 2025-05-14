package integer_test

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/ratel/keys/integer"
)

func TestT(t *testing.T) {
	for range 50 {
		number := frand.Bytes(integer.Len)
		v := integer.NewFrom(number)
		// log.I.S(number)
		buf := new(bytes.Buffer)
		v.Write(buf)
		buf2 := bytes.NewBuffer(buf.Bytes())
		v2 := &integer.T{} // or can use New(nil)
		el := v2.Read(buf2).(*integer.T)
		if el.Val != v.Val {
			t.Fatalf("expected %x got %x", v.Val, el.Val)
		}
		// log.I.S(el, v, v2)
	}
}
