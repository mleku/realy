package event

import (
	"bufio"
	"bytes"
	"testing"

	"realy.lol/event/examples"
)

func TestFromCompact(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, com by
	var err er
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(by, 0, len(b))
		c = append(c, b...)
		ea := New()
		if rem, err = ea.Unmarshal(b); chk.E(err) {
			t.Fatalf("error: %s", err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'", rem)
		}
		com = ea.MarshalCompact(com)
		mrsh := ea.Marshal(nil)
		eb := &T{}
		if rem, err = eb.UnmarshalCompact(com); chk.E(err) {
			t.Fatalf("error: %s", err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after FromCanonical: '%0x'", rem)
		}
		mrsh2 := eb.Marshal(nil)
		if !bytes.Equal(mrsh, mrsh2) {
			t.Fatalf("compact mismatch:\n\n%s\n%s", mrsh, mrsh2)
		}
		com, rem = com[:0], rem[:0]
	}
}
