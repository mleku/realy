package event

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"realy.lol/chk"
	"realy.lol/event/examples"
)

func TestTMarshalBinary_UnmarshalBinary(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out []byte
	var err error
	buf := new(bytes.Buffer)
	ea, eb := New(), New()
	now := time.Now()
	var counter int
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make([]byte, 0, len(b))
		c = append(c, b...)
		if rem, err = ea.Unmarshal(c); chk.E(err) {
			t.Fatal(err)
		}
		// log.I.F("len %d\n%s\n", len(b), ea.SerializeIndented())
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		ea.MarshalBinary(buf)
		// log.I.S(buf.Bytes())
		buf2 := bytes.NewBuffer(buf.Bytes())
		if err = eb.UnmarshalBinary(buf2); chk.E(err) {
			t.Fatal(err)
		}
		// log.I.F("len %d\n%s\n", len(b), eb.SerializeIndented())
		counter++
		out = out[:0]
		// break
	}
	t.Logf("unmarshaled json, marshaled binary, unmarshaled binary, %d events in %v av %v per event",
		counter, time.Since(now), time.Since(now)/time.Duration(counter))
}
