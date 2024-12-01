package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"

	"realy.lol/event/examples"
	"realy.lol/p256k"
)

func TestTMarshal_Unmarshal(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out by
	var err er
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(by, 0, len(b))
		c = append(c, b...)
		ea := New()
		if _, err = ea.UnmarshalJSON(b); chk.E(err) {
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
		out = out[:0]
	}
}

func TestT_CheckSignature(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out by
	var err er
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make(by, 0, len(b))
		c = append(c, b...)
		ea := New()
		if _, err = ea.UnmarshalJSON(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var valid bo
		if valid, err = ea.Verify(); chk.E(err) {
			t.Fatal(err)
		}
		if !valid {
			t.Fatalf("invalid signature\n%s", b)
		}
		out = out[:0]
	}
}

func TestT_SignWithSecKey(t *testing.T) {
	var err er
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var ev *T
	for _ = range 1000 {
		if ev, err = GenerateRandomTextNoteEvent(signer, 1000); chk.E(err) {
			t.Fatal(err)
		}
		var valid bo
		if valid, err = ev.Verify(); chk.E(err) {
			t.Fatal(err)
		}
		if !valid {
			b, _ := ev.MarshalJSON(nil)
			t.Fatalf("invalid signature\n%s", b)
		}
	}
}

// func TestBinaryEvents(t *testing.T) {
//	var err error
//	var ev, ev2 *T
//	_ = ev2
//	var orig by
//	b2, b3 := make(by, 0, 1_000_000), make(by, 0, 1_000_000)
//	j2, j3 := make(by, 0, 1_000_000), make(by, 0, 1_000_000)
//	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
//	buf := make(by, 1_000_000)
//	scanner.Buffer(buf, len(buf))
//	ev, ev2 = New(), New()
//	for scanner.Scan() {
//		orig = scanner.Bytes()
//		var cp by
//		cp = append(cp, orig...)
//		if orig, err = ev.UnmarshalJSON(orig); chk.E(err) {
//			t.Fatal(err)
//		}
//		if len(orig) > 0 {
//			t.Fatalf("remainder after end of event: %s", orig)
//		}
//		if b2, err = ev.MarshalBinary(b2); chk.E(err) {
//			t.Fatal(err)
//		}
//		// copy for verification
//		b3 = append(b3, b2...)
//		if b2, err = ev2.UnmarshalBinary(b2); chk.E(err) {
//			t.Fatal(err)
//		}
//		if j2, err = ev2.MarshalJSON(j2); chk.E(err) {
//			t.Fatal(err)
//		}
//		if len(b2) > 0 {
//			t.Fatalf("remainder after end of event: %s", orig)
//		}
//		// bytes should be identical to b3
//		if b2, err = ev2.MarshalBinary(b2); chk.E(err) {
//			es := err.Error()
//			if strings.Contains(es, "invalid length event ID in `a` tag:") {
//				err = nil
//				goto zero
//			}
//			log.E.Ln(es)
//		}
//		if !equals(b2, b3) {
//			// log.E.S(ev, ev2)
//			t.Fatalf("failed to remarshal\n%0x\n%0x",
//				b3, b2)
//		}
//	zero:
//		j2, j3 = j2[:0], j3[:0]
//		b2, b3 = b2[:0], b3[:0]
//	}
// }

func BenchmarkMarshalJSON(bb *testing.B) {
	bb.StopTimer()
	var i no
	var out by
	var err er
	evts := make([]*T, 0, 9999)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(by, 1_000_000)
	scanner.Buffer(buf, len(buf))
	for scanner.Scan() {
		b := scanner.Bytes()
		ea := New()
		if b, err = ea.UnmarshalJSON(b); chk.E(err) {
			bb.Fatal(err)
		}
		evts = append(evts, ea)
	}
	bb.ReportAllocs()
	var counter no
	out = out[:0]
	bb.StartTimer()
	for i = 0; i < bb.N; i++ {
		out, _ = evts[counter].MarshalJSON(out)
		out = out[:0]
		counter++
		if counter != len(evts) {
			counter = 0
		}
	}
}

func BenchmarkUnmarshalJSON(bb *testing.B) {
	var i no
	var err er
	evts := make([]*T, 9999)
	bb.ReportAllocs()
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(by, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var counter no
	for i = 0; i < bb.N; i++ {
		if !scanner.Scan() {
			scanner = bufio.NewScanner(bytes.NewBuffer(examples.Cache))
			scanner.Scan()
		}
		b := scanner.Bytes()
		ea := New()
		if b, err = ea.UnmarshalJSON(b); chk.E(err) {
			bb.Fatal(err)
		}
		evts[counter] = ea
		b = b[:0]
		if counter > 9999 {
			counter = 0
		}
	}
}

// func BenchmarkMarshalBinary(bb *testing.by) {
//	bb.StopTimer()
//	var i int
//	var out by
//	var err error
//	evts := make([]*T, 0, 9999)
//	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
//	buf := make(by, 1_000_000)
//	scanner.Buffer(buf, len(buf))
//	for scanner.Scan() {
//		b := scanner.Bytes()
//		ea := New()
//		if b, err = ea.UnmarshalJSON(b); chk.E(err) {
//			bb.Fatal(err)
//		}
//		evts = append(evts, ea)
//	}
//	var counter int
//	out = out[:0]
//	bb.ReportAllocs()
//	bb.StartTimer()
//	for i = 0; i < bb.N; i++ {
//		out, _ = evts[counter].MarshalBinary(out)
//		out = out[:0]
//		counter++
//		if counter != len(evts) {
//			counter = 0
//		}
//	}
// }
//
// func BenchmarkUnmarshalBinary(bb *testing.by) {
//	bb.StopTimer()
//	var i int
//	var out by
//	var err error
//	evts := make([]by, 0, 9999)
//	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
//	buf := make(by, 1_000_000)
//	scanner.Buffer(buf, len(buf))
//	for scanner.Scan() {
//		b := scanner.Bytes()
//		ea := New()
//		if b, err = ea.UnmarshalJSON(b); chk.E(err) {
//			bb.Fatal(err)
//		}
//		out = make(by, len(b))
//		out, _ = ea.MarshalBinary(out)
//		evts = append(evts, out)
//	}
//	bb.ReportAllocs()
//	var counter int
//	bb.StartTimer()
//	ev := New()
//	for i = 0; i < bb.N; i++ {
//		l := len(evts[counter])
//		b := make(by, l)
//		copy(b, evts[counter])
//		b, _ = ev.UnmarshalBinary(b)
//		out = out[:0]
//		counter++
//		if counter != len(evts) {
//			counter = 0
//		}
//	}
// }
