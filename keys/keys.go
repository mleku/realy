// Package keys is a set of helpers for generating and converting public/secret
// keys to hex and back to binary.
package keys

import (
	"bytes"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/p256k"
)

// GeneratePrivateKey - deprecated, use GenerateSecretKeyHex
var GeneratePrivateKey = func() string { return GenerateSecretKeyHex() }

// GenerateSecretKey creates a new secret key and returns the bytes of the secret.
func GenerateSecretKey() (skb []byte, err error) {
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	return
}

// GenerateSecretKeyHex generates a secret key and encodes the bytes as hex.
func GenerateSecretKeyHex() (sks string) {
	skb, err := GenerateSecretKey()
	if chk.E(err) {
		return
	}
	return hex.Enc(skb)
}

// GetPublicKeyHex generates a public key from a hex encoded secret key.
func GetPublicKeyHex(sk string) (pk string, err error) {
	var b []byte
	if b, err = hex.Dec(sk); chk.E(err) {
		return
	}
	signer := &p256k.Signer{}
	if err = signer.InitSec(b); chk.E(err) {
		return
	}

	return hex.Enc(signer.Pub()), nil
}

// SecretBytesToPubKeyHex generates a public key from secret key bytes.
func SecretBytesToPubKeyHex(skb []byte) (pk string, err error) {
	signer := &p256k.Signer{}
	if err = signer.InitSec(skb); chk.E(err) {
		return
	}
	return hex.Enc(signer.Pub()), nil
}

// IsValid32ByteHex checks that a hex string is a valid 32 bytes lower case hex encoded value as
// per nostr NIP-01 spec.
func IsValid32ByteHex[V []byte | string](pk V) bool {
	if bytes.Equal(bytes.ToLower([]byte(pk)), []byte(pk)) {
		return false
	}
	var err error
	dec := make([]byte, 32)
	if _, err = hex.DecBytes(dec, []byte(pk)); chk.E(err) {
	}
	return len(dec) == 32
}

// IsValidPublicKey checks that a hex encoded public key is a valid BIP-340 public key.
func IsValidPublicKey[V []byte | string](pk V) bool {
	v, _ := hex.Dec(string(pk))
	_, err := schnorr.ParsePubKey(v)
	return err == nil
}

// HexPubkeyToBytes decodes a pubkey from hex encoded string/bytes.
func HexPubkeyToBytes[V []byte | string](hpk V) (pkb []byte, err error) {
	return hex.DecAppend(nil, []byte(hpk))
}
