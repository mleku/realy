package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"
)

//go:embed test.jsonl
var example []byte

func TestFromWireCompact(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(example))
	var rem, com []byte
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make([]byte, 0, len(b))
		c = append(c, b...)
		ea := New()
		if rem, err = ea.Unmarshal(b); chk.E(err) {
			t.Fatalf("error: %s", err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'", rem)
		}
		com = ea.MarshalWireCompact(com)
		mrsh := ea.Marshal(nil)
		log.I.F("\n%s\n%s\n", com, mrsh)
		eb := &T{}
		if rem, err = eb.UnmarshalWireCompact(com); chk.E(err) {
			t.Fatalf("error: %s", err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after FromCanonical: '%s'", rem)
		}
		mrsh2 := eb.Marshal(nil)
		if !bytes.Equal(mrsh, mrsh2) {
			t.Fatalf("compact mismatch:\n\n%s\n%s", mrsh, mrsh2)
		}
		com, rem = com[:0], rem[:0]
	}
}
