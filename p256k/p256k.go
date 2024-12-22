//go:build cgo

package p256k

import (
	btcec "realy.lol/ec"
	realy "realy.lol/signer"
	"realy.lol/ec/schnorr"
)

func init() {
	log.T.Ln("using bitcoin/secp256k1 signature library")
}

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
	skb, pkb    by
}

var _ realy.I = &Signer{}

func (s *Signer) Generate() (err er) {
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

// InitSec loads a provided secret key into the signer. This also initializes
// a pubkey, as well as enabling creating signatures.
func (s *Signer) InitSec(skb by) (err er) {
	var cs *Sec
	var xp *XPublicKey
	var cp *PublicKey
	if s.pkb, cs, xp, cp, err = FromSecretBytes(skb); chk.E(err) {
		if err.Error() != "provided secret generates a public key with odd Y coordinate, fixed version returned" {
			log.E.Ln(err)
			return
		}
	}
	s.skb = skb
	s.SecretKey = &cs.Key
	s.PublicKey = xp.Key
	s.ECPublicKey = cp.Key
	s.BTCECSec, _ = btcec.PrivKeyFromBytes(s.skb)
	return
}

// InitPub initializes a signer to do verification. This can either be a 257 bit
// 33 byte key with 2 or 3 prefix or an x-only pubkey that is the same without
// the prefix.
func (s *Signer) InitPub(pub by) (err er) {
	var up *Pub
	if up, err = PubFromBytes(pub); chk.E(err) {
		return
	}
	s.PublicKey = &up.Key
	s.pkb = up.PubB()
	return
}

func (s *Signer) Sec() (b by)   { return s.skb }
func (s *Signer) Pub() (b by)   { return s.pkb[len(s.pkb)-schnorr.PubKeyBytesLen:] }
func (s *Signer) ECPub() (b by) { return s.pkb }

func (s *Signer) Sign(msg by) (sig by, err er) {
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

func (s *Signer) Verify(msg, sig by) (valid bo, err er) {
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
func (s *Signer) ECDH(xkb by) (secret by, err er) {
	var pubKey *btcec.PublicKey
	k2 := append(by{2}, xkb...)
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
