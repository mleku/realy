package auth

import (
	"testing"

	"realy.lol/p256k"
)

func TestCreateUnsigned(t *testing.T) {
	var err error
	signer := new(p256k.Signer)
	if err = signer.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var ok bool
	const relayURL = "wss://example.com"
	for range 100 {
		challenge := GenerateChallenge()
		ev := CreateUnsigned(signer.Pub(), challenge, relayURL)
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
