package text

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/hex"
	"realy.mleku.dev/sha256"
)

func TestUnmarshalHexArray(t *testing.T) {
	var ha [][]byte
	h := make([]byte, sha256.Size)
	frand.Read(h)
	var dst []byte
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
	var ha2 [][]byte
	var rem []byte
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
		if !bytes.Equal(ha[i], ha2[i]) {
			t.Fatalf("failed to unmarshal at element %d; got %x, expected %x",
				i, ha[i], ha2[i])
		}
	}
}
