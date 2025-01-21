package btcec

import (
	ec "realy.lol/ec"
	"realy.lol/ec/schnorr"
	"realy.lol/ec/secp256k1"
	"realy.lol/signer"
)

type Signer struct {
	SecretKey *secp256k1.SecretKey
	PublicKey *secp256k1.PublicKey
	pkb, skb  by
}

var _ signer.I = &Signer{}

func (s *Signer) Generate() (err er) {
	for {
		if s.SecretKey, err = ec.NewSecretKey(); chk.E(err) {
			return
		}
		s.skb = s.SecretKey.Serialize()
		s.PublicKey = s.SecretKey.PubKey()
		s.pkb = s.PublicKey.SerializeCompressed()
		if s.pkb[0] == 2 {
			s.skb = s.SecretKey.Serialize()
			break
		} else {
			s.Negate()
			s.pkb = s.PublicKey.SerializeCompressed()
			break
		}
	}
	return
}

func (s *Signer) InitSec(sec by) (err er) {
	if len(sec) != secp256k1.SecKeyBytesLen {
		err = errorf.E("sec key must be %d bytes", secp256k1.SecKeyBytesLen)
		return
	}
	s.SecretKey = secp256k1.SecKeyFromBytes(sec)
	s.PublicKey = s.SecretKey.PubKey()
	s.pkb = s.SecretKey.PubKey().SerializeCompressed()
	return
}

func (s *Signer) InitPub(pub by) (err er) {
	if s.PublicKey, err = schnorr.ParsePubKey(pub); chk.E(err) {
		return
	}
	s.pkb = pub
	return
}

func (s *Signer) Sec() (b by)   { return s.skb }
func (s *Signer) Pub() (b by)   { return s.pkb[1:] }
func (s *Signer) ECPub() (b by) { return s.pkb }

func (s *Signer) Sign(msg by) (sig by, err er) {
	if s.SecretKey == nil {
		err = errorf.E("btcec: Signer not initialized")
		return
	}
	var si *schnorr.Signature
	if si, err = schnorr.Sign(s.SecretKey, msg); chk.E(err) {
		return
	}
	sig = si.Serialize()
	return
}

func (s *Signer) Verify(msg, sig by) (valid bo, err er) {
	if s.PublicKey == nil {
		err = errorf.E("btcec: PubKey not initialized")
		return
	}
	var si *schnorr.Signature
	if si, err = schnorr.ParseSignature(sig); chk.D(err) {
		err = errorf.E("failed to parse signature:\n%d %s\n%v", len(sig),
			sig, err)
		return
	}
	valid = si.Verify(msg, s.PublicKey)
	return
}

func (s *Signer) Zero() { s.SecretKey.Key.Zero() }

func (s *Signer) ECDH(pubkeyBytes by) (secret by, err er) {
	var pub *secp256k1.PublicKey
	// Note the public key must be even, if the secret key derives an odd compressed pubkey ECDH will fail in that
	// direction.
	if pub, err = secp256k1.ParsePubKey(append(by{0x02},
		pubkeyBytes...)); chk.E(err) {
		return
	}
	secret = ec.GenerateSharedSecret(s.SecretKey, pub)
	return
}

func (s *Signer) Negate() {
	s.SecretKey.Key = *s.SecretKey.Key.Negate()
	s.skb = s.SecretKey.Serialize()
	s.PublicKey = s.SecretKey.PubKey()
	s.pkb = s.PublicKey.SerializeCompressed()
}

type Keygen struct {
	Signer
}

func (k *Keygen) Generate() (pubBytes by, err er) {
	if k.Signer.SecretKey, err = ec.NewSecretKey(); chk.E(err) {
		return
	}
	k.Signer.PublicKey = k.SecretKey.PubKey()
	k.Signer.pkb = k.PublicKey.SerializeCompressed()
	pubBytes = k.Signer.pkb
	return
}

func (k *Keygen) Negate() {
	k.Signer.Negate()
	return
}

func (k *Keygen) KeyPairBytes() (secBytes, cmprPubBytes by) {
	return k.Signer.SecretKey.Serialize(), k.Signer.PublicKey.SerializeCompressed()
}
