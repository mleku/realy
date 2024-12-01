package filter

import (
	"testing"
)

func TestT_MarshalUnmarshal(t *testing.T) {
	var err er
	const bufLen = 4000000
	dst := make(by, 0, bufLen)
	dst1 := make(by, 0, bufLen)
	dst2 := make(by, 0, bufLen)
	for _ = range 20 {
		f := New()
		if f, err = GenFilter(); chk.E(err) {
			t.Fatal(err)
		}
		if dst, err = f.MarshalJSON(dst); chk.E(err) {
			t.Fatal(err)
		}
		dst1 = append(dst1, dst...)
		// now unmarshal
		var rem by
		fa := New()
		if rem, err = fa.UnmarshalJSON(dst); chk.E(err) {
			t.Fatalf("unmarshal error: %v\n%s\n%s", err, dst, rem)
		}
		dst2, _ = fa.MarshalJSON(nil)
		if !equals(dst1, dst2) {
			t.Fatalf("marshal error: %v\n%s\n%s", err, dst1, dst2)
		}
		dst, dst1, dst2 = dst[:0], dst1[:0], dst2[:0]
	}
}
