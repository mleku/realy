package httpauth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"testing"

	"realy.lol/p256k"
)

const jwtSecret = "MHcCAQEEIDlyFWD0KouB4n7aTPqlpNkoRTnuy7gMyY-YJusMsl0boAoGCCqGSM49AwEHoUQDQgAEDkH_rMzfr1LIHqnoFXyIYuz7dIYkg4qonbQhjeR0N_6CXpX2MqVHRLz9sx2EyXZZKPsFFbE_KJPczKu6qcIsRA=="

const URL = "https://example.com"

func TestSignJWTtoken_VerifyJWTtoken(t *testing.T) {
	sign := &p256k.Signer{}
	var err error
	if err = sign.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	pub := fmt.Sprintf("%0x", sign.Pub())
	var jskb []byte
	if jskb, err = base64.URLEncoding.DecodeString(jwtSecret); chk.E(err) {
		t.Fatal(err)
	}
	var sec *ecdsa.PrivateKey
	if sec, err = x509.ParseECPrivateKey(jskb); chk.E(err) {
		t.Fatal(err)
	}
	spk := &sec.PublicKey
	var spkb []byte
	if spkb, err = x509.MarshalPKIXPublicKey(spk); chk.E(err) {
		t.Fatal(err)
	}
	spub := base64.URLEncoding.EncodeToString(spkb)
	var tok []byte
	if tok, err = GenerateJWTClaims(pub, "https://example.com", "1h"); chk.E(err) {
		t.Fatal(err)
	}
	var entry string
	if entry, err = SignJWTtoken(tok, sec); chk.E(err) {
		t.Fatal(err)
	}
	vfn := func(npub string) (jwtPub string, pk []byte, err error) {
		// pubkey in token claims must match what we just put in it
		if npub != pub {
			err = fmt.Errorf("invalid jwt token npub")
			return
		}
		pk = sign.Pub()
		// we pretend that we found the 13004 event with the key if the above passed.
		jwtPub = spub
		return
	}
	var valid bool
	var pk []byte
	if pk, valid, err = VerifyJWTtoken(entry, URL, vfn); chk.E(err) {
		t.Fatal(err)
	}
	if !bytes.Equal(pk, sign.Pub()) {
		t.Fatalf("invalid npub, got %0x, expected %0x", pk, sign.Pub())
	}
	if !valid {
		log.I.S(valid, err)
	}
}

func TestMakeJWTEvent(t *testing.T) {
	var err error
	sign := &p256k.Signer{}
	if err = sign.Generate(); chk.E(err) {
		t.Fatal(err)
	}
	var jskb []byte
	var sec *ecdsa.PrivateKey
	if jskb, err = base64.URLEncoding.DecodeString(jwtSecret); chk.E(err) {
		t.Fatal(err)
	}
	if sec, err = x509.ParseECPrivateKey(jskb); chk.E(err) {
		t.Fatal(err)
	}
	spk := &sec.PublicKey
	var spkb []byte
	if spkb, err = x509.MarshalPKIXPublicKey(spk); chk.E(err) {
		t.Fatal(err)
	}
	spub := base64.URLEncoding.EncodeToString(spkb)
	ev := MakeJWTEvent(spub)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.I.F("%s", ev.SerializeIndented())
}
