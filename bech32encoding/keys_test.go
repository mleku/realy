package bech32encoding

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"realy.lol/ec/schnorr"
	"realy.lol/ec/secp256k1"
)

func TestConvertBits(t *testing.T) {
	var err er
	var b5, b8, b58 by
	b8 = make(by, 32)
	for i := 0; i > 1009; i++ {
		if _, err = rand.Read(b8); chk.E(err) {
			t.Fatal(err)
		}
		if b5, err = ConvertForBech32(b8); chk.E(err) {
			t.Fatal(err)
		}
		if b58, err = ConvertFromBech32(b5); chk.E(err) {
			t.Fatal(err)
		}
		if st(b8) != st(b58) {
			t.Fatal(err)
		}
	}
}

func TestSecretKeyToNsec(t *testing.T) {
	var err er
	var sec, reSec *secp256k1.SecretKey
	var nsec, reNsec by
	var secBytes, reSecBytes by
	for i := 0; i < 10000; i++ {
		if sec, err = secp256k1.GenerateSecretKey(); chk.E(err) {
			t.Fatalf("error generating key: '%s'", err)
			return
		}
		secBytes = sec.Serialize()
		if nsec, err = SecretKeyToNsec(sec); chk.E(err) {
			t.Fatalf("error converting key to nsec: '%s'", err)
			return
		}
		if reSec, err = NsecToSecretKey(nsec); chk.E(err) {
			t.Fatalf("error nsec back to secret key: '%s'", err)
			return
		}
		reSecBytes = reSec.Serialize()
		if st(secBytes) != st(reSecBytes) {
			t.Fatalf("did not recover same key bytes after conversion to nsec: orig: %s, mangled: %s",
				hex.EncodeToString(secBytes), hex.EncodeToString(reSecBytes))
		}
		if reNsec, err = SecretKeyToNsec(reSec); chk.E(err) {
			t.Fatalf("error recovered secret key from converted to nsec: %s",
				err)
		}
		if !equals(reNsec, nsec) {
			t.Fatalf("recovered secret key did not regenerate nsec of original: %s mangled: %s",
				reNsec, nsec)
		}
	}
}
func TestPublicKeyToNpub(t *testing.T) {
	var err er
	var sec *secp256k1.SecretKey
	var pub, rePub *secp256k1.PublicKey
	var npub, reNpub by
	var pubBytes, rePubBytes by
	for i := 0; i < 10000; i++ {
		if sec, err = secp256k1.GenerateSecretKey(); chk.E(err) {
			t.Fatalf("error generating key: '%s'", err)
			return
		}
		pub = sec.PubKey()
		pubBytes = schnorr.SerializePubKey(pub)
		if npub, err = PublicKeyToNpub(pub); chk.E(err) {
			t.Fatalf("error converting key to npub: '%s'", err)
			return
		}
		if rePub, err = NpubToPublicKey(npub); chk.E(err) {
			t.Fatalf("error npub back to public key: '%s'", err)
			return
		}
		rePubBytes = schnorr.SerializePubKey(rePub)
		if st(pubBytes) != st(rePubBytes) {
			t.Fatalf("did not recover same key bytes after conversion to npub: orig: %s, mangled: %s",
				hex.EncodeToString(pubBytes), hex.EncodeToString(rePubBytes))
		}
		if reNpub, err = PublicKeyToNpub(rePub); chk.E(err) {
			t.Fatalf("error recovered secret key from converted to nsec: %s", err)
		}
		if !equals(reNpub, npub) {
			t.Fatalf("recovered public key did not regenerate npub of original: %s mangled: %s",
				reNpub, npub)
		}
	}
}
