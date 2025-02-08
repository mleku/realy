package filters

import (
	"bytes"
	"testing"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err error
	dst := make([]byte, 0, 4000000)
	dst1 := make([]byte, 0, len(dst))
	dst2 := make([]byte, 0, len(dst))
	for _ = range 1000 {
		var f1 *T
		if f1, err = GenFilters(5); chk.E(err) {
			t.Fatal(err)
		}
		// now unmarshal
		dst = f1.Marshal(dst)
		dst1 = append(dst1, dst...)
		// now unmarshal
		var rem []byte
		f2 := New()
		if rem, err = f2.Unmarshal(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2 = f2.Marshal(dst2)
		if !bytes.Equal(dst1, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst1, dst2)
		}
		dst, dst1, dst2 = dst[:0], dst1[:0], dst2[:0]
	}
}
