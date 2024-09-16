//go:build !btcec

package p256k_test

import (
	"bufio"
	"bytes"
	"testing"

	"mleku.dev/ec/schnorr"
	"mleku.dev/event"
	"mleku.dev/event/examples"
	"mleku.dev/p256k"
	"mleku.dev/sha256"
)

func TestVerify(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	for scanner.Scan() {
		var valid bool
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.UnmarshalJSON(b); chk.E(err) {
			t.Errorf("failed to marshal\n%s", b)
		} else {
			if valid, err = ev.Verify(); chk.E(err) || !valid {
				t.Errorf("btcec: invalid signature\n%s", b)
				continue
			}
		}
		id := ev.GetIDBytes()
		if len(id) != sha256.Size {
			t.Errorf("id should be 32 bytes, got %d", len(id))
			continue
		}
		if err = p256k.VerifyFromBytes(id, ev.Sig, ev.PubKey); chk.E(err) {
			t.Error(err)
			continue
		}
		evs = append(evs, ev)
	}
}

func TestSign(t *testing.T) {
	evs := make([]*event.T, 0, 10000)
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	buf := make(B, 1_000_000)
	scanner.Buffer(buf, len(buf))
	var err error
	var sec1 *p256k.Sec
	var pub1 *p256k.XPublicKey
	var pb B
	if _, pb, sec1, pub1, _, err = p256k.Generate(); chk.E(err) {
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
	sig := make(B, schnorr.SignatureSize)
	for _, ev := range evs {
		ev.PubKey = pb
		var uid *p256k.Uchar
		if uid, err = p256k.Msg(ev.GetIDBytes()); chk.E(err) {
			t.Fatal(err)
		}
		if sig, err = p256k.Sign(uid, sec1.Sec()); chk.E(err) {
			t.Fatal(err)
		}
		ev.Sig = sig
		var usig *p256k.Uchar
		if usig, err = p256k.Sig(sig); chk.E(err) {
			t.Fatal(err)
		}
		if !p256k.Verify(uid, usig, pub1.Key) {
			t.Errorf("invalid signature")
		}
	}
	p256k.Zero(&sec1.Key)
}
