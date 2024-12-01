package keys

import (
	"strings"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/p256k"
)

// GeneratePrivateKey - deprecated, use GenerateSecretKeyHex
var GeneratePrivateKey = func() st { return GenerateSecretKeyHex() }

func GenerateSecretKey() (skb by, err er) {
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	return
}

func GenerateSecretKeyHex() (sks st) {
	skb, err := GenerateSecretKey()
	if chk.E(err) {
		return
	}
	return hex.Enc(skb)
}

func GetPublicKeyHex(sk st) (pk st, err er) {
	var b by
	if b, err = hex.Dec(sk); chk.E(err) {
		return
	}
	signer := &p256k.Signer{}
	if err = signer.InitSec(b); chk.E(err) {
		return
	}

	return hex.Enc(signer.Pub()), nil
}

func SecretBytesToPubKeyHex(skb by) (pk st, err er) {
	signer := &p256k.Signer{}
	if err = signer.InitSec(skb); chk.E(err) {
		return
	}
	return hex.Enc(signer.Pub()), nil
}

func IsValid32ByteHex(pk st) bo {
	if strings.ToLower(pk) != pk {
		return false
	}
	dec, _ := hex.Dec(pk)
	return len(dec) == 32
}

func IsValidPublicKey(pk st) bo {
	v, _ := hex.Dec(pk)
	_, err := schnorr.ParsePubKey(v)
	return err == nil
}

func HexPubkeyToBytes[V by | st](hpk V) (pkb by, err er) {
	return hex.DecAppend(nil, by(hpk))
}
