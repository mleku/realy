package pubkey

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/ec/schnorr"
)

func TestT(t *testing.T) {
	for _ = range 10000000 {
		fakePubkeyBytes := frand.Bytes(schnorr.PubKeyBytesLen)
		v, err := New(fakePubkeyBytes)
		if chk.E(err) {
			t.FailNow()
		}
		buf := new(bytes.Buffer)
		v.Write(buf)
		buf2 := bytes.NewBuffer(buf.Bytes())
		v2, _ := New()
		el := v2.Read(buf2).(*T)
		if bytes.Compare(el.Val, v.Val) != 0 {
			t.Fatalf("expected %x got %x", v.Val, el.Val)
		}
	}
}
