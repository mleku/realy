package bech32encoding

import (
	btcec "realy.lol/ec"
	"realy.lol/ec/bech32"
	"realy.lol/ec/schnorr"
	"realy.lol/ec/secp256k1"
	"realy.lol/hex"
)

const (
	// MinKeyStringLen is 56 because Bech32 needs 52 characters plus 4 for the HRP,
	// any string shorter than this cannot be a nostr key.
	MinKeyStringLen = 56
	HexKeyLen       = 64
	Bech32HRPLen    = 4
)

var (
	SecHRP = by("nsec")
	PubHRP = by("npub")
)

// ConvertForBech32 performs the bit expansion required for encoding into Bech32.
func ConvertForBech32(b8 by) (b5 by, err er) { return bech32.ConvertBits(b8, 8,
	5, true) }

// ConvertFromBech32 collapses together the bit expanded 5 bit numbers encoded in bech32.
func ConvertFromBech32(b5 by) (b8 by, err er) { return bech32.ConvertBits(b5, 5,
	8, true) }

// SecretKeyToNsec encodes an secp256k1 secret key as a Bech32 string (nsec).
func SecretKeyToNsec(sk *secp256k1.SecretKey) (encoded by, err er) {
	var b5 by
	if b5, err = ConvertForBech32(sk.Serialize()); chk.E(err) {
		return
	}
	return bech32.Encode(SecHRP, b5)
}

// PublicKeyToNpub encodes a public key as a bech32 string (npub).
func PublicKeyToNpub(pk *secp256k1.PublicKey) (encoded by, err er) {
	var bits5 by
	pubKeyBytes := schnorr.SerializePubKey(pk)
	if bits5, err = ConvertForBech32(pubKeyBytes); chk.E(err) {
		return
	}
	return bech32.Encode(PubHRP, bits5)
}

// NsecToSecretKey decodes a nostr secret key (nsec) and returns the secp256k1
// secret key.
func NsecToSecretKey(encoded by) (sk *secp256k1.SecretKey, err er) {
	var b5, b8, hrp by
	if hrp, b5, err = bech32.Decode(encoded); chk.E(err) {
		return
	}
	if !equals(hrp, SecHRP) {
		err = log.E.Err("wrong human readable part, got '%s' want '%s'",
			hrp, SecHRP)
		return
	}
	if b8, err = ConvertFromBech32(b5); chk.E(err) {
		return
	}
	sk = secp256k1.SecKeyFromBytes(b8)
	return
}

// NpubToPublicKey decodes an nostr public key (npub) and returns an secp256k1
// public key.
func NpubToPublicKey(encoded by) (pk *secp256k1.PublicKey, err er) {
	var b5, b8, hrp by
	if hrp, b5, err = bech32.Decode(encoded); chk.E(err) {
		err = log.E.Err("ERROR: '%s'", err)
		return
	}
	if !equals(hrp, PubHRP) {
		err = log.E.Err("wrong human readable part, got '%s' want '%s'",
			hrp, PubHRP)
		return
	}
	if b8, err = ConvertFromBech32(b5); chk.E(err) {
		return
	}

	return schnorr.ParsePubKey(b8[:32])
}

// HexToPublicKey decodes a string that should be a 64 character long hex
// encoded public key into a btcec.PublicKey that can be used to verify a
// signature or encode to Bech32.
func HexToPublicKey(pk st) (p *btcec.PublicKey, err er) {
	if len(pk) != HexKeyLen {
		err = log.E.Err("secret key is %d bytes, must be %d", len(pk),
			HexKeyLen)
		return
	}
	var pb by
	if pb, err = hex.Dec(pk); chk.D(err) {
		return
	}
	if p, err = schnorr.ParsePubKey(pb); chk.D(err) {
		return
	}
	return
}

// HexToSecretKey decodes a string that should be a 64 character long hex
// encoded public key into a btcec.PublicKey that can be used to verify a
// signature or encode to Bech32.
func HexToSecretKey(sk by) (s *btcec.SecretKey, err er) {
	if len(sk) != HexKeyLen {
		err = log.E.Err("secret key is %d bytes, must be %d", len(sk),
			HexKeyLen)
		return
	}
	pb := make(by, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, sk); chk.D(err) {
		return
	}
	if s = secp256k1.SecKeyFromBytes(pb); chk.D(err) {
		return
	}
	return
}

func HexToNpub(publicKeyHex by) (s by, err er) {
	b := make(by, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(b, publicKeyHex); chk.D(err) {
		err = log.E.Err("failed to decode public key hex: %w", err)
		return
	}
	var bits5 by
	if bits5, err = bech32.ConvertBits(b, 8, 5, true); chk.D(err) {
		return nil, err
	}
	return bech32.Encode(NpubHRP, bits5)
}

func BinToNpub(b by) (npub by, err er) {
	var bits5 by
	if bits5, err = bech32.Convert8to5(b, true); chk.D(err) {
		return nil, err
	}
	return bech32.Encode(NpubHRP, bits5)
}

// HexToNsec converts a hex encoded secret key to a bech32 encoded nsec.
func HexToNsec(sk by) (nsec by, err er) {
	var s *btcec.SecretKey
	if s, err = HexToSecretKey(sk); chk.E(err) {
		return
	}
	if nsec, err = SecretKeyToNsec(s); chk.E(err) {
		return
	}
	return
}

// BinToNsec converts a binary secret key to a bech32 encoded nsec.
func BinToNsec(sk by) (nsec by, err er) {
	var s *btcec.SecretKey
	s, _ = btcec.SecKeyFromBytes(sk)
	if nsec, err = SecretKeyToNsec(s); chk.E(err) {
		return
	}
	return
}

// SecretKeyToHex converts a secret key to the hex encoding.
func SecretKeyToHex(sk *btcec.SecretKey) (hexSec by) {
	hex.EncBytes(hexSec, sk.Serialize())
	return
}

func NsecToHex(nsec by) (hexSec by, err er) {
	var sk *secp256k1.SecretKey
	if sk, err = NsecToSecretKey(nsec); chk.E(err) {
		return
	}
	hexSec = SecretKeyToHex(sk)
	return
}
