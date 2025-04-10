//go:build cgo

package p256k

import "C"
import (
	btcec "realy.mleku.dev/ec"
	"realy.mleku.dev/ec/secp256k1"
	realy "realy.mleku.dev/signer"
)

func init() {
	log.T.Ln("using bitcoin/secp256k1 signature library")
}

// Signer implements the signer.I interface.
//
// Either the Sec or Pub must be populated, the former is for generating
// signatures, the latter is for verifying them.
//
// When using this library only for verification, a constructor that converts
// from bytes to PubKey is needed prior to calling Verify.
type Signer struct {
	// SecretKey is the secret key.
	SecretKey *SecKey
	// PublicKey is the public key.
	PublicKey *PubKey
	// BTCECSec is needed for ECDH as currently the CGO bindings don't include it
	BTCECSec *btcec.SecretKey
	skb, pkb []byte
}

var _ realy.I = &Signer{}

// Generate a new Signer key pair using the CGO bindings to libsecp256k1
func (s *Signer) Generate() (err error) {
	var cs *Sec
	var cx *XPublicKey
	if s.skb, s.pkb, cs, cx, err = Generate(); chk.E(err) {
		return
	}
	s.SecretKey = &cs.Key
	s.PublicKey = cx.Key
	s.BTCECSec, _ = btcec.PrivKeyFromBytes(s.skb)
	return
}

func (s *Signer) InitSec(skb []byte) (err error) {
	var cs *Sec
	var cx *XPublicKey
	// var cp *PublicKey
	if s.pkb, cs, cx, err = FromSecretBytes(skb); chk.E(err) {
		if err.Error() != "provided secret generates a public key with odd Y coordinate, fixed version returned" {
			log.E.Ln(err)
			return
		}
	}
	s.skb = skb
	s.SecretKey = &cs.Key
	s.PublicKey = cx.Key
	// s.ECPublicKey = cp.Key
	// needed for ecdh
	s.BTCECSec, _ = btcec.PrivKeyFromBytes(s.skb)
	return
}

func (s *Signer) InitPub(pub []byte) (err error) {
	var up *Pub
	if up, err = PubFromBytes(pub); chk.E(err) {
		return
	}
	s.PublicKey = &up.Key
	s.pkb = up.PubB()
	return
}

func (s *Signer) Sec() (b []byte) { return s.skb }
func (s *Signer) Pub() (b []byte) { return s.pkb }

// func (s *Signer) ECPub() (b []byte) { return s.pkb }

func (s *Signer) Sign(msg []byte) (sig []byte, err error) {
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

func (s *Signer) Verify(msg, sig []byte) (valid bool, err error) {
	if s.PublicKey == nil {
		err = errorf.E("p256k: Pubkey not initialized")
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

func (s *Signer) ECDH(pubkeyBytes []byte) (secret []byte, err error) {
	var pub *secp256k1.PublicKey
	if pub, err = secp256k1.ParsePubKey(append([]byte{0x02}, pubkeyBytes...)); chk.E(err) {
		return
	}
	secret = btcec.GenerateSharedSecret(s.BTCECSec, pub)
	return
}

func (s *Signer) Zero() { Zero(s.SecretKey) }
