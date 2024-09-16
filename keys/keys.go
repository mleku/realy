package keys

import (
	"strings"

	btcec "mleku.dev/ec"
	"mleku.dev/ec/schnorr"
	"mleku.dev/hex"
	"mleku.dev/p256k"
)

var GeneratePrivateKey = func() B { return GenerateSecretKeyHex() }

func GenerateSecretKeyHex() (sks B) {
	var err E
	var skb B
	signer := &p256k.Signer{}
	if err = signer.Generate(); chk.E(err) {
		return
	}
	skb = signer.Sec()
	sks = B(hex.Enc(skb))
	return
}

func GetPublicKeyHex(sk S) (S, E) {
	b, err := hex.Dec(sk)
	if chk.E(err) {
		return "", err
	}
	_, pk := btcec.PrivKeyFromBytes(b)
	return hex.Enc(schnorr.SerializePubKey(pk)), nil
}

func SecretBytesToPubKeyHex(skb B) (pk S, err E) {
	_, pkk := btcec.SecKeyFromBytes(skb)
	return hex.Enc(schnorr.SerializePubKey(pkk)), nil
}

func SecretToPubKeyBytes(skb B) (pk B, err E) {
	_, pkk := btcec.SecKeyFromBytes(skb)
	return schnorr.SerializePubKey(pkk), nil
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
