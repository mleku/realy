package bech32encoding

import (
	"bytes"

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
	// HexKeyLen is the length of a nostr key in hexadecimal.
	HexKeyLen = 64
	// Bech32HRPLen is the length of the standard nostr keys, nsec and npub.
	Bech32HRPLen = 4
)

var (
	// SecHRP is the standard Human Readable Prefix (HRP) for a nostr secret key in bech32 encoding - nsec
	SecHRP = []byte("nsec")
	// PubHRP is the standard Human Readable Prefix (HRP) for a nostr public key in bech32 encoding - nsec
	PubHRP = []byte("npub")
)

// ConvertForBech32 performs the bit expansion required for encoding into Bech32.
func ConvertForBech32(b8 []byte) (b5 []byte, err error) { return bech32.ConvertBits(b8, 8, 5, true) }

// ConvertFromBech32 collapses together the bit expanded 5 bit numbers encoded in bech32.
func ConvertFromBech32(b5 []byte) (b8 []byte, err error) { return bech32.ConvertBits(b5, 5, 8, true) }

// SecretKeyToNsec encodes an secp256k1 secret key as a Bech32 string (nsec).
func SecretKeyToNsec(sk *secp256k1.SecretKey) (encoded []byte, err error) {
	var b5 []byte
	if b5, err = ConvertForBech32(sk.Serialize()); chk.E(err) {
		return
	}
	return bech32.Encode(SecHRP, b5)
}

// PublicKeyToNpub encodes a public key as a bech32 string (npub).
func PublicKeyToNpub(pk *secp256k1.PublicKey) (encoded []byte, err error) {
	var bits5 []byte
	pubKeyBytes := schnorr.SerializePubKey(pk)
	if bits5, err = ConvertForBech32(pubKeyBytes); chk.E(err) {
		return
	}
	return bech32.Encode(PubHRP, bits5)
}

// NsecToSecretKey decodes a nostr secret key (nsec) and returns the secp256k1
// secret key.
func NsecToSecretKey(encoded []byte) (sk *secp256k1.SecretKey, err error) {
	var b8 []byte
	if b8, err = NsecToBytes(encoded); chk.E(err) {
		return
	}
	sk = secp256k1.SecKeyFromBytes(b8)
	return
}

// NsecToBytes converts a nostr bech32 encoded secret key to raw bytes.
func NsecToBytes(encoded []byte) (sk []byte, err error) {
	var b5, hrp []byte
	if hrp, b5, err = bech32.Decode(encoded); chk.E(err) {
		return
	}
	if !bytes.Equal(hrp, SecHRP) {
		err = log.E.Err("wrong human readable part, got '%s' want '%s'",
			hrp, SecHRP)
		return
	}
	if sk, err = ConvertFromBech32(b5); chk.E(err) {
		return
	}
	sk = sk[:secp256k1.SecKeyBytesLen]
	return
}

// NpubToBytes converts a bech32 encoded public key to raw bytes.
func NpubToBytes(encoded []byte) (pk []byte, err error) {
	var b5, hrp []byte
	if hrp, b5, err = bech32.Decode(encoded); chk.E(err) {
		return
	}
	if !bytes.Equal(hrp, PubHRP) {
		err = log.E.Err("wrong human readable part, got '%s' want '%s'",
			hrp, SecHRP)
		return
	}
	if pk, err = ConvertFromBech32(b5); chk.E(err) {
		return
	}
	pk = pk[:schnorr.PubKeyBytesLen]
	return
}

// NpubToPublicKey decodes an nostr public key (npub) and returns an secp256k1
// public key.
func NpubToPublicKey(encoded []byte) (pk *secp256k1.PublicKey, err error) {
	var b5, b8, hrp []byte
	if hrp, b5, err = bech32.Decode(encoded); chk.E(err) {
		err = log.E.Err("ERROR: '%s'", err)
		return
	}
	if !bytes.Equal(hrp, PubHRP) {
		err = log.E.Err("wrong human readable part, got '%s' want '%s'",
			hrp, PubHRP)
		return
	}
	if b8, err = ConvertFromBech32(b5); chk.E(err) {
		return
	}

	return schnorr.ParsePubKey(b8[:schnorr.PubKeyBytesLen])
}

// HexToPublicKey decodes a string that should be a 64 character long hex
// encoded public key into a btcec.PublicKey that can be used to verify a
// signature or encode to Bech32.
func HexToPublicKey(pk string) (p *btcec.PublicKey, err error) {
	if len(pk) != HexKeyLen {
		err = log.E.Err("secret key is %d bytes, must be %d", len(pk),
			HexKeyLen)
		return
	}
	var pb []byte
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
func HexToSecretKey(sk []byte) (s *btcec.SecretKey, err error) {
	if len(sk) != HexKeyLen {
		err = log.E.Err("secret key is %d bytes, must be %d", len(sk),
			HexKeyLen)
		return
	}
	pb := make([]byte, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, sk); chk.D(err) {
		return
	}
	if s = secp256k1.SecKeyFromBytes(pb); chk.D(err) {
		return
	}
	return
}

// HexToNpub converts a raw 64 character hex encoded public key (as used in
// standard nostr json events) to a bech32 encoded npub.
func HexToNpub(publicKeyHex []byte) (s []byte, err error) {
	b := make([]byte, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(b, publicKeyHex); chk.D(err) {
		err = log.E.Err("failed to decode public key hex: %w", err)
		return
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(b, 8, 5, true); chk.D(err) {
		return nil, err
	}
	return bech32.Encode(NpubHRP, bits5)
}

// BinToNpub converts a raw 32 byte public key to nostr bech32 encoded npub.
func BinToNpub(b []byte) (s []byte, err error) {
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(b, 8, 5, true); chk.D(err) {
		return nil, err
	}
	return bech32.Encode(NpubHRP, bits5)
}

// HexToNsec converts a hex encoded secret key to a bech32 encoded nsec.
func HexToNsec(sk []byte) (nsec []byte, err error) {
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
func BinToNsec(sk []byte) (nsec []byte, err error) {
	var s *btcec.SecretKey
	s, _ = btcec.SecKeyFromBytes(sk)
	if nsec, err = SecretKeyToNsec(s); chk.E(err) {
		return
	}
	return
}

// SecretKeyToHex converts a secret key to the hex encoding.
func SecretKeyToHex(sk *btcec.SecretKey) (hexSec []byte) {
	hex.EncBytes(hexSec, sk.Serialize())
	return
}

// NsecToHex converts a bech32 encoded nostr secret key to a raw hexadecimal
// string.
func NsecToHex(nsec []byte) (hexSec []byte, err error) {
	var sk *secp256k1.SecretKey
	if sk, err = NsecToSecretKey(nsec); chk.E(err) {
		return
	}
	hexSec = SecretKeyToHex(sk)
	return
}
