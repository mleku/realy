package authenvelope

import (
	"bytes"
	"testing"

	"realy.lol/auth"
	"realy.lol/chk"
	"realy.lol/envelopes"
	"realy.lol/p256k"
)

const relayURL = "wss://example.com"

func TestAuth(t *testing.T) {
	var err error
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var b1, b2, b3, b4 []byte
	for _ = range 1000 {
		ch := auth.GenerateChallenge()
		chal := Challenge{Challenge: ch}
		b1 = chal.Marshal(b1)
		oChal := make([]byte, len(b1))
		copy(oChal, b1)
		var rem []byte
		var l string
		if l, b1 = envelopes.Identify(b1); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		c2 := NewChallenge()
		if rem, err = c2.Unmarshal(b1); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("remainder should be empty\n%s", rem)
		}
		if !bytes.Equal(chal.Challenge, c2.Challenge) {
			t.Fatalf("challenge mismatch\n%s\n%s",
				chal.Challenge, c2.Challenge)
		}
		b2 = c2.Marshal(b2)
		if !bytes.Equal(oChal, b2) {
			t.Fatalf("challenge mismatch\n%s\n%s", oChal, b2)
		}
		resp := Response{Event: auth.CreateUnsigned(signer.Pub(), ch,
			relayURL)}
		if err = resp.Event.Sign(signer); chk.E(err) {
			t.Fatal(err)
		}
		b3 = resp.Marshal(b3)
		oResp := make([]byte, len(b3))
		copy(oResp, b3)
		if l, b3 = envelopes.Identify(b3); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		r2 := NewResponse()
		if _, err = r2.Unmarshal(b3); chk.E(err) {
			t.Fatal(err)
		}
		b4 = r2.Marshal(b4)
		if !bytes.Equal(oResp, b4) {
			t.Fatalf("challenge mismatch\n%s\n%s", oResp, b4)
		}
		b1, b2, b3, b4 = b1[:0], b2[:0], b3[:0], b4[:0]
		oChal, oResp = oChal[:0], oResp[:0]
	}
}
