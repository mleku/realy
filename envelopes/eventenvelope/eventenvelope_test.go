package eventenvelope

import (
	"bufio"
	"bytes"
	"testing"

	"realy.lol/envelopes"
	"realy.lol/event"
	"realy.lol/event/examples"
	"realy.lol/subscription"
)

func TestSubmission(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		rem = rem[:0]
		ea := NewSubmissionWith(ev)
		if rem, err = ea.MarshalJSON(rem); chk.E(err) {
			t.Fatal(err)
		}
		c = append(c, rem...)
		var l string
		if l, rem, err = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.UnmarshalJSON(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		if out, err = ea.MarshalJSON(out); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		c, out = c[:0], out[:0]
	}
}

func TestResult(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out B
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
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
		if rem, err = ea.MarshalJSON(rem); chk.E(err) {
			t.Fatal(err)
		}
		c = append(c, rem...)
		var l string
		if l, rem, err = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.UnmarshalJSON(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		if out, err = ea.MarshalJSON(out); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		rem, c, out = rem[:0], c[:0], out[:0]
	}
}
