package event

import (
	"bufio"
	"bytes"
	"testing"

	"realy.lol/event/examples"
)

func TestFromCanonical(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out, can []byte
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make([]byte, 0, len(b))
		c = append(c, b...)
		ea := New()
		if rem, err = ea.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'", rem)
		}
		can = ea.ToCanonical(can)
		ea.Sig = ea.Sig[:0]
		mrsh := ea.Marshal(nil)
		eb := &T{}
		if rem, err = eb.FromCanonical(can); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after FromCanonical: '%s'", rem)
		}
		mrsh2 := eb.Marshal(nil)
		if !bytes.Equal(mrsh, mrsh2) {
			t.Fatalf("canonical mismatch:\n\n%s\n%s", mrsh, mrsh2)
		}
		out, can, rem = out[:0], can[:0], rem[:0]
	}
}
