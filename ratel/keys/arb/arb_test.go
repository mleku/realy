package arb

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"
)

func TestT(t *testing.T) {
	randomBytes := frand.Bytes(frand.Intn(128))
	v := New(randomBytes)
	buf := new(bytes.Buffer)
	v.Write(buf)
	randomCopy := make([]byte, len(randomBytes))
	buf2 := bytes.NewBuffer(buf.Bytes())
	v2 := New(randomCopy)
	el := v2.Read(buf2).(*T)
	if bytes.Compare(el.Val, v.Val) != 0 {
		t.Fatalf("expected %x got %x", v.Val, el.Val)
	}
}
