package keys

import (
	"strings"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/p256k"
)

// GeneratePrivateKey - deprecated, use GenerateSecretKeyHex
var GeneratePrivateKey = func() S { return GenerateSecretKeyHex() }

func GenerateSecretKey() (skb B, err E) {
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	return
}

func GenerateSecretKeyHex() (sks S) {
	skb, err := GenerateSecretKey()
	if chk.E(err) {
		return
	}
	return hex.Enc(skb)
}

func GetPublicKeyHex(sk S) (pk S, err E) {
	var b B
	if b, err = hex.Dec(sk); chk.E(err) {
		return
	}
	signer := &p256k.Signer{}
	if err = signer.InitSec(b); chk.E(err) {
		return
	}

	return hex.Enc(signer.Pub()), nil
}

func SecretBytesToPubKeyHex(skb B) (pk S, err E) {
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

func HexPubkeyToBytes[V B | S](hpk V) (pkb B, err E) {
	return hex.DecAppend(nil, B(hpk))
}
