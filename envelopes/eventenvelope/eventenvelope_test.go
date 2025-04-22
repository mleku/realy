package eventenvelope

import (
	"bufio"
	"bytes"
	"testing"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/event"
	"realy.mleku.dev/event/examples"
	"realy.mleku.dev/subscription"
)

func TestSubmission(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out []byte
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		rem = rem[:0]
		ea := NewSubmissionWith(ev)
		rem = ea.Marshal(rem)
		c = append(c, rem...)
		var l string
		if l, rem = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.Unmarshal(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		out = ea.Marshal(out)
		if !bytes.Equal(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		c, out = c[:0], out[:0]
	}
}

func TestResult(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out []byte
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var ea *Result
		if ea, err = NewResultWith(subscription.NewStd().String(), ev); chk.E(err) {
			t.Fatal(err)
		}
		rem = ea.Marshal(rem)
		c = append(c, rem...)
		var l string
		if l, rem = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.Unmarshal(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		out = ea.Marshal(out)
		if !bytes.Equal(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		rem, c, out = rem[:0], c[:0], out[:0]
	}
}
