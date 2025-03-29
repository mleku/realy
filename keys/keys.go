// Package keys is a set of helpers for generating and converting public/secret
// keys to hex and back to binary.
package keys

import (
	"strings"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/p256k"
)

// GeneratePrivateKey - deprecated, use GenerateSecretKeyHex
var GeneratePrivateKey = func() string { return GenerateSecretKeyHex() }

func GenerateSecretKey() (skb []byte, err error) {
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	return
}

func GenerateSecretKeyHex() (sks string) {
	skb, err := GenerateSecretKey()
	if chk.E(err) {
		return
	}
	return hex.Enc(skb)
}

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

func SecretBytesToPubKeyHex(skb []byte) (pk string, err error) {
	signer := &p256k.Signer{}
	if err = signer.InitSec(skb); chk.E(err) {
		return
	}
	return hex.Enc(signer.Pub()), nil
}

func IsValid32ByteHex(pk string) bool {
	if strings.ToLower(pk) != pk {
		return false
	}
	dec, _ := hex.Dec(pk)
	return len(dec) == 32
}

func IsValidPublicKey(pk string) bool {
	v, _ := hex.Dec(pk)
	_, err := schnorr.ParsePubKey(v)
	return err == nil
}

func HexPubkeyToBytes[V []byte | string](hpk V) (pkb []byte, err error) {
	return hex.DecAppend(nil, []byte(hpk))
}
