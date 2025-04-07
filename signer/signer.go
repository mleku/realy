// Package signer defines interfaces for management of signatures, used to
// abstract the signature algorithm from the usage.
package signer

// I is an interface for a key pair for signing, created to abstract between a CGO fast BIP-340
// signature library and the slower btcec library.
type I interface {
	// Generate creates a fresh new key pair from system entropy, and ensures it is even (so
	// ECDH works).
	Generate() (err error)
	// InitSec initialises the secret (signing) key from the raw bytes, and also
	// derives the public key because it can.
	InitSec(sec []byte) (err error)
	// InitPub initializes the public (verification) key from raw bytes, this is
	// expected to be an x-only 32 byte pubkey.
	InitPub(pub []byte) (err error)
	// Sec returns the secret key bytes.
	Sec() []byte
	// Pub returns the public key bytes (x-only schnorr pubkey).
	Pub() []byte
	// Sign creates a signature using the stored secret key.
	Sign(msg []byte) (sig []byte, err error)
	// Verify checks a message hash and signature match the stored public key.
	Verify(msg, sig []byte) (valid bool, err error)
	// Zero wipes the secret key to prevent memory leaks.
	Zero()
	// ECDH returns a shared secret derived using Elliptic Curve Diffie-Hellman on
	// the I secret and provided pubkey.
	ECDH(pub []byte) (secret []byte, err error)
}

// Gen is an interface for nostr BIP-340 key generation.
type Gen interface {
	// Generate gathers entropy and derives pubkey bytes for matching, this returns the 33 byte
	// compressed form for checking the oddness of the Y coordinate.
	Generate() (pubBytes []byte, err error)
	// Negate flips the public key Y coordinate between odd and even.
	Negate()
	// KeyPairBytes returns the raw bytes of the secret and public key, this returns the 32 byte
	// X-only pubkey.
	KeyPairBytes() (secBytes, cmprPubBytes []byte)
}
