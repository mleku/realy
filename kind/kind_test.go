package kind

import (
	"testing"

	"lukechampine.com/frand"

	"realy.lol/chk"
)

func TestMarshalUnmarshal(t *testing.T) {
	var err error
	k := make([]*T, 1000000)
	for i := range k {
		k[i] = New(uint16(frand.Intn(65535)))
	}
	mk := make([][]byte, len(k))
	for i := range mk {
		mk[i] = make([]byte, 0, 5) // 16 bits max 65535 = 5 characters
	}
	for i := range k {
		mk[i] = k[i].Marshal(mk[i])
	}
	k2 := make([]*T, len(k))
	for i := range k2 {
		k2[i] = New(0)
	}
	for i := range k2 {
		var r []byte
		if r, err = k2[i].Unmarshal(mk[i]); chk.E(err) {
			t.Fatal(err)
		}
		if len(r) != 0 {
			t.Fatalf("remainder after unmarshal: '%s'", r)
		}
	}
}
