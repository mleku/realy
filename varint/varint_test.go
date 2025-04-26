package varint

import (
	"bytes"
	"math"
	"testing"

	"lukechampine.com/frand"

	"realy.lol/chk"
)

func TestEncode_Decode(t *testing.T) {
	var v uint64
	for range 10000000 {
		v = uint64(frand.Intn(math.MaxInt64))
		buf1 := new(bytes.Buffer)
		Encode(buf1, v)
		buf2 := bytes.NewBuffer(buf1.Bytes())
		u, err := Decode(buf2)
		if chk.E(err) {
			t.Fatal(err)
		}
		if u != v {
			t.Fatalf("expected %d got %d", v, u)
		}

	}
}
