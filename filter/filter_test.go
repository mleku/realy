package filter

import (
	"bytes"
	"testing"

	"realy.lol/chk"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err error
	const bufLen = 4000000
	dst := make([]byte, 0, bufLen)
	dst1 := make([]byte, 0, bufLen)
	dst2 := make([]byte, 0, bufLen)
	for _ = range 20 {
		f := New()
		if f, err = GenFilter(); chk.E(err) {
			t.Fatal(err)
		}
		dst = f.Marshal(dst)
		dst1 = append(dst1, dst...)
		// now unmarshal
		var rem []byte
		fa := New()
		if rem, err = fa.Unmarshal(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2 = fa.Marshal(nil)
		if !bytes.Equal(dst1, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst1, dst2)
		}
		dst, dst1, dst2 = dst[:0], dst1[:0], dst2[:0]
	}
}
