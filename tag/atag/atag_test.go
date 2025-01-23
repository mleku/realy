package atag

import (
	"testing"
	"realy.lol/kind"
	"lukechampine.com/frand"
	"math"
	"realy.lol/ec/schnorr"
	"realy.lol/hex"
)

func TestT_Marshal_Unmarshal(t *testing.T) {
	k := kind.New(frand.Intn(math.MaxUint16))
	pk := make(by, schnorr.PubKeyBytesLen)
	frand.Read(pk)
	d := make(by, frand.Intn(10)+3)
	frand.Read(d)
	var dtag st
	dtag = hex.Enc(d)
	t1 := &T{
		Kind:   k,
		PubKey: pk,
		DTag:   by(dtag),
	}
	b1 := t1.Marshal(nil)
	log.I.F("%s", b1)
	t2 := &T{}
	var r by
	var err er
	if r, err = t2.Unmarshal(b1); chk.E(err) {
		t.Fatal(err)
	}
	if len(r) > 0 {
		log.I.S(r)
		t.Fatalf("remainder")
	}
	b2 := t2.Marshal(nil)
	if !equals(b1, b2) {
		t.Fatalf("failed to re-marshal back original")
	}
}
