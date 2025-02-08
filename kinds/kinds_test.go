package kinds

import (
	"testing"

	"lukechampine.com/frand"

	"realy.lol/kind"
)

func TestUnmarshalKindsArray(t *testing.T) {
	k := &T{make([]*kind.T, 100)}
	for i := range k.K {
		k.K[i] = kind.New(uint16(frand.Intn(65535)))
	}
	var dst []byte
	var err error
	if dst = k.Marshal(dst); chk.E(err) {
		t.Fatal(err)
	}
	k2 := &T{}
	var rem []byte
	if rem, err = k2.Unmarshal(dst); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		t.Fatalf("failed to unmarshal, remnant afterwards '%s'", rem)
	}
	for i := range k.K {
		if *k.K[i] != *k2.K[i] {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, k.K[i], k2.K[i])
		}
	}
}
