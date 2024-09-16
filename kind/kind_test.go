package kind

import (
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalUnmarshal(t *testing.T) {
	var err error
	k := make([]*T, 1000000)
	for i := range k {
		k[i] = New(uint16(frand.Intn(65535)))
	}
	mk := make([]B, len(k))
	for i := range mk {
		mk[i] = make(B, 0, 5) // 16 bits max 65535 = 5 characters
	}
	for i := range k {
		if mk[i], err = k[i].MarshalJSON(mk[i]); chk.E(err) {
			t.Fatal(err)
		}
	}
	k2 := make([]*T, len(k))
	for i := range k2 {
		k2[i] = New(0)
	}
	for i := range k2 {
		var r B
		if r, err = k2[i].UnmarshalJSON(mk[i]); chk.E(err) {
			t.Fatal(err)
		}
		if len(r) != 0 {
			t.Fatalf("remainder after unmarshal: '%s'", r)
		}
	}
}
