//go:build cgo

package p256k_test

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/event/examples"
	"realy.lol/p256k"
	realy "realy.lol/signer"
)

func TestSigner_Generate(t *testing.T) {
	for _ = range 10000 {
		var err er
		signer := &p256k.Signer{}
		var skb by
		if err = signer.Generate(); chk.E(err) {
			t.Fatal(err)
		}
		skb = signer.Sec()
		if err = signer.InitSec(skb); chk.E(err) {
			t.Fatal(err)
		}
	}
}

func TestSignerVerify(t *testing.T) {
	// evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(by, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err er
	signer := &p256k.Signer{}
	for scanner.Scan() {
		var valid bo
		b := scanner.Bytes()
		bc := make(by, 0, len(b))
		bc = append(bc, b...)
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		} else {
			if valid, err = ev.Verify(); chk.T(err) || !valid {
				t.Errorf("invalid signature\n%s", bc)
				continue
			}
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		if err = signer.InitPub(ev.PubKey); chk.T(err) {
			t.Errorf("failed to init pub key: %s\n%0x", err, ev.PubKey)
			continue
		}
		if valid, err = signer.Verify(id, ev.Sig); chk.E(err) {
			t.Errorf("failed to verify: %s\n%0x", err, ev.ID)
			continue
		}
		if !valid {
			t.Errorf("invalid signature for\npub %0x\neid %0x\nsig %0x\n%s",
				ev.PubKey, id, ev.Sig, bc)
			continue
		}
		// fmt.Printf("%s\n", bc)
		// evs = append(evs, ev)
	}
}

func TestSignerSign(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(by, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err er
	signer := &p256k.Signer{}
	var skb, pkb by
	if skb, pkb, _, _, _, err = p256k.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	if err = signer.InitSec(skb); chk.E(err) {
		t.Fatal(err)
	}
	verifier := &p256k.Signer{}
	if err = verifier.InitPub(pkb[1:]); chk.E(err) {
		t.Fatal(err)
	}
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		}
		evs = append(evs, ev)
	}
	var valid bo
	sig := make(by, schnorr.SignatureSize)
	for _, ev := range evs {
		ev.PubKey = pkb
		id := ev.GetIDBytes()
		if sig, err = signer.Sign(id); chk.E(err) {
			t.Errorf("failed to sign: %s\n%0x", err, id)
		}
		if valid, err = verifier.Verify(id, sig); chk.E(err) {
			t.Errorf("failed to verify: %s\n%0x", err, id)
		}
		if !valid {
			t.Errorf("invalid signature")
		}
	}
	signer.Zero()
}

func TestECDH(t *testing.T) {
	n := time.Now()
	var err er
	var s1, s2 realy.I
	var counter no
	const total = 100
	for _ = range total {
		s1, s2 = &p256k.Signer{}, &p256k.Signer{}
		if err = s1.Generate(); chk.E(err) {
			t.Fatal(err)
		}
		if err = s2.Generate(); chk.E(err) {
			t.Fatal(err)
		}
		for _ = range total {
			var secret1, secret2 by
			if secret1, err = s1.ECDH(s2.Pub()); chk.E(err) {
				t.Fatal(err)
			}
			if secret2, err = s2.ECDH(s1.Pub()); chk.E(err) {
				t.Fatal(err)
			}
			if !equals(secret1, secret2) {
				counter++
				t.Errorf("ECDH generation failed to work in both directions, %x %x", secret1,
					secret2)
			}
		}
	}
	a := time.Now()
	duration := a.Sub(n)
	log.I.Ln("errors", counter, "total", total, "time", duration, "time/op",
		no(duration/total),
		"ops/sec", no(time.Second)/no(duration/total))
}
