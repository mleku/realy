package fullpubkey

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/sha256"
)

func TestT(t *testing.T) {
	pk := frand.Bytes(sha256.Size)
	v := New(pk)
	buf := new(bytes.Buffer)
	v.Write(buf)
	buf2 := bytes.NewBuffer(buf.Bytes())
	v2 := New()
	el := v2.Read(buf2).(*T)
	if bytes.Compare(el.Val, v.Val) != 0 {
		t.Fatalf("expected %x got %x", v.Val, el.Val)
	}
}
