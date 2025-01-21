package auth

import (
	"testing"

	"realy.lol/p256k"
)

func TestCreateUnsigned(t *testing.T) {
	var err er
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var ok bo
	const relayURL = "wss://example.com"
	for _ = range 100 {
		challenge := GenerateChallenge()
		ev := CreateUnsigned(challenge, relayURL)
		if err = ev.Sign(signer); chk.E(err) {
			t.Fatal(err)
		}
		if ok, err = Validate(ev, challenge, relayURL); chk.E(err) {
			t.Fatal(err)
		}
		if !ok {
			bb := ev.Marshal(nil)
			t.Fatalf("failed to validate auth event\n%s", bb)
		}
	}
}
