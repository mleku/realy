package text

import (
	"testing"

	"lukechampine.com/frand"
	"mleku.dev/hex"
	"mleku.dev/sha256"
)

func TestUnmarshalHexArray(t *testing.T) {
	var ha []B
	h := make(B, sha256.Size)
	frand.Read(h)
	var dst B
	for _ = range 20 {
		hh := sha256.Sum256(h)
		h = hh[:]
		ha = append(ha, h)
	}
	dst = append(dst, '[')
	for i := range ha {
		dst = AppendQuote(dst, ha[i], hex.EncAppend)
		if i != len(ha)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, ']')
	var ha2 []B
	var rem B
	var err error
	if ha2, rem, err = UnmarshalHexArray(dst, 32); chk.E(err) {
		t.Fatal(err)
	}
	if len(ha2) != len(ha) {
		t.Fatalf("failed to unmarshal, got %d fields, expected %d", len(ha2),
			len(ha))
	}
	if len(rem) > 0 {
		t.Fatalf("failed to unmarshal, remnant afterwards '%s'", rem)
	}
	for i := range ha2 {
		if !equals(ha[i], ha2[i]) {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, ha[i], ha2[i])
		}
	}
}
