package kinds

import (
	"testing"

	"lukechampine.com/frand"
	"mleku.dev/kind"
)

func TestUnmarshalKindsArray(t *testing.T) {
	k := &T{make([]*kind.T, 100)}
	for i := range k.K {
		k.K[i] = kind.New(uint16(frand.Intn(65535)))
	}
	var dst B
	var err error
	if dst, err = k.MarshalJSON(dst); chk.E(err) {
		t.Fatal(err)
	}
	k2 := &T{}
	var rem B
	if rem, err = k2.UnmarshalJSON(dst); chk.E(err) {
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
