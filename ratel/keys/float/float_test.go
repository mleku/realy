package float_test

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/log"
	"realy.lol/ratel/keys/float"
)

func TestT(t *testing.T) {
	for range 50 {
		fakefloat := frand.Bytes(float.Len)
		v := float.NewFrom(fakefloat)
		log.I.S(fakefloat)
		buf := new(bytes.Buffer)
		v.Write(buf)
		buf2 := bytes.NewBuffer(buf.Bytes())
		v2 := &float.S{} // or can use New(nil)
		el := v2.Read(buf2).(*float.S)
		if el.Val != v.Val {
			t.Fatalf("expected %x got %x", v.Val, el.Val)
		}
		log.I.S(el, v, v2)
	}
}
