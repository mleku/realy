package authenvelope

import (
	"testing"

	"mleku.dev/auth"
	"mleku.dev/envelopes"
	"mleku.dev/p256k"
)

const relayURL = "wss://example.com"

func TestAuth(t *testing.T) {
	var err error
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var b1, b2, b3, b4 B
	for _ = range 1000 {
		ch := auth.GenerateChallenge()
		chal := Challenge{Challenge: ch}
		if b1, err = chal.MarshalJSON(b1); chk.E(err) {
			t.Fatal(err)
		}
		oChal := make(B, len(b1))
		copy(oChal, b1)
		var rem B
		var l string
		if l, b1, err = envelopes.Identify(b1); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		c2 := NewChallenge()
		if rem, err = c2.UnmarshalJSON(b1); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("remainder should be empty\n%s", rem)
		}
		if !equals(chal.Challenge, c2.Challenge) {
			t.Fatalf("challenge mismatch\n%s\n%s",
				chal.Challenge, c2.Challenge)
		}
		if b2, err = c2.MarshalJSON(b2); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(oChal, b2) {
			t.Fatalf("challenge mismatch\n%s\n%s", oChal, b2)
		}
		resp := Response{Event: auth.CreateUnsigned(signer.Pub(), ch,
			relayURL)}
		if err = resp.Event.Sign(signer); chk.E(err) {
			t.Fatal(err)
		}
		if b3, err = resp.MarshalJSON(b3); chk.E(err) {
			t.Fatal(err)
		}
		oResp := make(B, len(b3))
		copy(oResp, b3)
		if l, b3, err = envelopes.Identify(b3); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		r2 := NewResponse()
		if _, err = r2.UnmarshalJSON(b3); chk.E(err) {
			t.Fatal(err)
		}
		if b4, err = r2.MarshalJSON(b4); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(oResp, b4) {
			t.Fatalf("challenge mismatch\n%s\n%s", oResp, b4)
		}
		b1, b2, b3, b4 = b1[:0], b2[:0], b3[:0], b4[:0]
		oChal, oResp = oChal[:0], oResp[:0]
	}
}
