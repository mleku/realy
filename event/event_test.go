package event

import (
	"bufio"
	"bytes"
	_ "embed"
	"testing"

	"realy.lol/chk"
	"realy.lol/event/examples"
	"realy.lol/p256k"
)

func TestTMarshal_Unmarshal(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out []byte
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
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		if out = ea.Marshal(out); chk.E(err) {
			t.Fatal(err)
		}
		if !bytes.Equal(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		out = out[:0]
	}
}

func TestT_CheckSignature(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var rem, out []byte
	var err error
	for scanner.Scan() {
		b := scanner.Bytes()
		c := make([]byte, 0, len(b))
		c = append(c, b...)
		ea := New()
		if _, err = ea.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var valid bool
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
	var err error
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var ev *T
	for _ = range 1000 {
		if ev, err = GenerateRandomTextNoteEvent(signer, 1000); chk.E(err) {
			t.Fatal(err)
		}
		var valid bool
		if valid, err = ev.Verify(); chk.E(err) {
			t.Fatal(err)
		}
		if !valid {
			b := ev.Marshal(nil)
			t.Fatalf("invalid signature\n%s", b)
		}
	}
}

func BenchmarkMarshal(bb *testing.B) {
	bb.StopTimer()
	var i int
	var out []byte
	var err error
	evts := make([]*T, 0, 9999)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make([]byte, 1_000_000)
	scanner.Buffer(buf, len(buf))
	for scanner.Scan() {
		b := scanner.Bytes()
		ea := New()
		if b, err = ea.Unmarshal(b); chk.E(err) {
			bb.Fatal(err)
		}
		evts = append(evts, ea)
	}
	bb.ReportAllocs()
	var counter int
	out = out[:0]
	bb.StartTimer()
	for i = 0; i < bb.N; i++ {
		out = evts[counter].Marshal(out)
		out = out[:0]
		counter++
		if counter != len(evts) {
			counter = 0
		}
	}
}

func BenchmarkUnmarshal(bb *testing.B) {
	var i int
	var err error
	evts := make([]*T, 9999)
	bb.ReportAllocs()
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make([]byte, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var counter int
	for i = 0; i < bb.N; i++ {
		if !scanner.Scan() {
			scanner = bufio.NewScanner(bytes.NewBuffer(examples.Cache))
			scanner.Scan()
		}
		b := scanner.Bytes()
		ea := New()
		if b, err = ea.Unmarshal(b); chk.E(err) {
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
//		if b, err = ea.Unmarshal(b); chk.E(err) {
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
//		if b, err = ea.Unmarshal(b); chk.E(err) {
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
