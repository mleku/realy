//go:build cgo

package p256k

import (
	btcec "realy.lol/ec"
	realy "realy.lol/signer"
)

// Signer implements the nostr.I interface.
//
// Either the Sec or Pub must be populated, the former is for generating
// signatures, the latter is for verifying them.
//
// When using this library only for verification, a constructor that converts
// from bytes to PubKey is needed prior to calling Verify.
type Signer struct {
	SecretKey   *SecKey
	PublicKey   *PubKey
	ECPublicKey *ECPubKey // not sure what this is useful for yet.
	BTCECSec    *btcec.SecretKey
	skb, pkb    B
}

var _ realy.I = &Signer{}

func (s *Signer) Generate() (err E) {
	var cs *Sec
	var cx *XPublicKey
	var cp *PublicKey
	if s.skb, s.pkb, cs, cx, cp, err = Generate(); chk.E(err) {
		return
	}
	s.SecretKey = &cs.Key
	s.PublicKey = cx.Key
	s.ECPublicKey = cp.Key
	s.BTCECSec, _ = btcec.PrivKeyFromBytes(s.skb)
	return
}

func (s *Signer) InitSec(skb B) (err error) {
	var cs *Sec
	var cx *XPublicKey
	var cp *PublicKey
	if s.pkb, cs, cx, cp, err = FromSecretBytes(skb); chk.E(err) {
		if err.Error() != "provided secret generates a public key with odd Y coordinate, fixed version returned" {
			log.E.Ln(err)
			return
		}
	}
	s.skb = skb
	s.SecretKey = &cs.Key
	s.PublicKey = cx.Key
	s.ECPublicKey = cp.Key
	s.BTCECSec, _ = btcec.PrivKeyFromBytes(s.skb)
	return
}

func (s *Signer) InitPub(pub B) (err error) {
	var up *Pub
	if up, err = PubFromBytes(pub); chk.E(err) {
		return
	}
	s.PublicKey = &up.Key
	s.pkb = up.PubB()
	return
}

func (s *Signer) Sec() (b B)   { return s.skb }
func (s *Signer) Pub() (b B)   { return s.pkb[1:] }
func (s *Signer) ECPub() (b B) { return s.pkb }

func (s *Signer) Sign(msg B) (sig B, err error) {
	if s.SecretKey == nil {
		err = errorf.E("p256k: I secret not initialized")
		return
	}
	u := ToUchar(msg)
	if sig, err = Sign(u, s.SecretKey); chk.E(err) {
		return
	}
	return
}

func (s *Signer) Verify(msg, sig B) (valid bool, err error) {
	if s.PublicKey == nil {
		err = errorf.E("p256k: PubKey not initialized")
		return
	}
	var uMsg, uSig *Uchar
	if uMsg, err = Msg(msg); chk.E(err) {
		return
	}
	if uSig, err = Sig(sig); chk.E(err) {
		return
	}
	valid = Verify(uMsg, uSig, s.PublicKey)
	if !valid {
		err = errorf.E("p256k: invalid signature")
	}
	return
}

func (s *Signer) Zero() { Zero(s.SecretKey) }
func (s *Signer) ECDH(xkb B) (secret B, err error) {
	var pubKey *btcec.PublicKey
	k2 := append(B{2}, xkb...)
	if pubKey, err = btcec.ParsePubKey(k2); chk.E(err) {
		err = errorf.E("error parsing receiver public key '%0x': %w", k2, err)
		return
	}
	secret = btcec.GenerateSharedSecret(s.BTCECSec, pubKey)
	return
}

func (s *Signer) Negate() {
	Negate(s.skb)
	chk.E(s.InitSec(s.skb))
}
