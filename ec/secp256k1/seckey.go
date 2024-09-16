// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2023 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package secp256k1

import (
	"crypto/rand"
	"io"
)

// SecretKey provides facilities for working with secp256k1 secret keys within
// this package and includes functionality such as serializing and parsing them
// as well as computing their associated public key.
type SecretKey struct {
	Key ModNScalar
}

// PrivateKey is a secret key.
//
// Deprecated: use SecretKey - secret = one person; private = two or more (you don't share secret keys!)
type PrivateKey = SecretKey

// NewSecretKey instantiates a new secret key from a scalar encoded as a big integer.
func NewSecretKey(key *ModNScalar) *SecretKey { return &SecretKey{Key: *key} }

// NewPrivateKey instantiates a new secret key from a scalar encoded as a big integer.
//
// Deprecated: use NewSecretKey - secret = one person; private = two or more (you don't share secret keys!)
var NewPrivateKey = NewSecretKey

// SecKeyFromBytes returns a secret based on the provided byte slice which is
// interpreted as an unsigned 256-bit big-endian integer in the range [0, N-1],
// where N is the order of the curve.
//
// WARNING: This means passing a slice with more than 32 bytes is truncated and
// that truncated value is reduced modulo N.  Further, 0 is not a valid secret
// key.  It is up to the caller to provide a value in the appropriate range of
// [1, N-1].  Failure to do so will either result in an invalid secret key or
// potentially weak secret keys that have bias that could be exploited.
//
// This function primarily exists to provide a mechanism for converting
// serialized secret keys that are already known to be good.
//
// Typically, callers should make use of GenerateSecretKey or
// GenerateSecretKeyFromRand when creating secret keys since they properly
// handle generation of appropriate values.
func SecKeyFromBytes(secKeyBytes []byte) *SecretKey {
	var secKey SecretKey
	secKey.Key.SetByteSlice(secKeyBytes)
	return &secKey
}

var PrivKeyFromBytes = SecKeyFromBytes

// generateSecretKey generates and returns a new secret key that is suitable
// for use with secp256k1 using the provided reader as a source of entropy.  The
// provided reader must be a source of cryptographically secure randomness to
// avoid weak secret keys.
func generateSecretKey(rand io.Reader) (*SecretKey, error) {
	// The group order is close enough to 2^256 that there is only roughly a 1
	// in 2^128 chance of generating an invalid secret key, so this loop will
	// virtually never run more than a single iteration in practice.
	var key SecretKey
	var b32 [32]byte
	for valid := false; !valid; {
		if _, err := io.ReadFull(rand, b32[:]); err != nil {
			return nil, err
		}
		// The secret key is only valid when it is in the range [1, N-1], where
		// N is the order of the curve.
		overflow := key.Key.SetBytes(&b32)
		valid = (key.Key.IsZeroBit() | overflow) == 0
	}
	zeroArray32(&b32)
	return &key, nil
}

// GenerateSecretKey generates and returns a new cryptographically secure secret key that is suitable for use with
// secp256k1.
func GenerateSecretKey() (*SecretKey, error) {
	return generateSecretKey(rand.Reader)
}

// GeneratePrivateKey generates and returns a new cryptographically secure secret key that is suitable for use with
// secp256k1.
//
// Deprecated: use NewSecretKey - secret = one person; private = two or more (you don't share secret keys!)
var GeneratePrivateKey = GenerateSecretKey

// GenerateSecretKeyFromRand generates a secret key that is suitable for use with secp256k1 using the provided reader as
// a source of entropy. The provided reader must be a source of cryptographically secure randomness, such as
// [crypto/rand.Reader], to avoid weak secret keys.
func GenerateSecretKeyFromRand(rand io.Reader) (*SecretKey, error) {
	return generateSecretKey(rand)
}

// GeneratePrivateKeyFromRand generates a secret key that is suitable for use with secp256k1 using the provided reader as
// a source of entropy. The provided reader must be a source of cryptographically secure randomness, such as
// [crypto/rand.Reader], to avoid weak secret keys.
//
// Deprecated: use GenerateSecretKeyFromRand - secret = one person; private = two or more (you don't share secret keys!)
var GeneratePrivateKeyFromRand = GenerateSecretKeyFromRand

// PubKey computes and returns the public key corresponding to this secret key.
func (p *SecretKey) PubKey() *PublicKey {
	var result JacobianPoint
	ScalarBaseMultNonConst(&p.Key, &result)
	result.ToAffine()
	return NewPublicKey(&result.X, &result.Y)
}

// Zero manually clears the memory associated with the secret key.  This can be
// used to explicitly clear key material from memory for enhanced security
// against memory scraping.
func (p *SecretKey) Zero() { p.Key.Zero() }

// SecKeyBytesLen defines the length in bytes of a serialized secret key.
const SecKeyBytesLen = 32

// Serialize returns the secret key as a 256-bit big-endian binary-encoded
// number, padded to a length of 32 bytes.
func (p *SecretKey) Serialize() []byte {
	var secKeyBytes [SecKeyBytesLen]byte
	p.Key.PutBytes(&secKeyBytes)
	return secKeyBytes[:]
}
