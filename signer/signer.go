package signer

type I interface {
	// Generate creates a fresh new key pair from system entropy, and ensures it
	// is even (so ECDH works).
	Generate() (err er)
	// InitSec initialises the secret (signing) key from the raw bytes, and also
	// derives the public key because it can.
	InitSec(sec by) (err er)
	// InitPub initializes the public (verification) key from raw bytes.
	InitPub(pub by) (err er)
	// Sec returns the secret key bytes.
	Sec() by
	// Pub returns the public key bytes (x-only schnorr pubkey).
	Pub() by
	// ECPub returns the public key bytes (33 byte ecdsa pubkey). The first byte is always 2 due
	// to ECDH and X-only keys.
	ECPub() by
	// Sign creates a signature using the stored secret key.
	Sign(msg by) (sig by, err er)
	// Verify checks a message hash and signature match the stored public key.
	Verify(msg, sig by) (valid bo, err er)
	// Zero wipes the secret key to prevent memory leaks.
	Zero()
	// ECDH returns a shared secret derived using Elliptic Curve Diffie-Hellman
	// on the secret and provided pubkey.
	ECDH(pub by) (secret by, err er)
	// Negate flips the secret key to change between odd and even compressed
	// public key.
	Negate()
}

// Gen is an interface for nostr BIP-340 key generation.
type Gen interface {
	// Generate gathers entropy and derives pubkey bytes for matching, this returns the 33 byte
	// compressed form for checking the oddness of the Y coordinate.
	Generate() (pubBytes by, err er)
	// Negate flips the public key Y coordinate between odd and even.
	Negate()
	// KeyPairBytes returns the raw bytes of the secret and public key, this returns the 32 byte
	// X-only pubkey.
	KeyPairBytes() (secBytes, cmprPubBytes by)
}
