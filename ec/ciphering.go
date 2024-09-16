// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcec

import (
	"mleku.dev/ec/secp256k1"
)

// GenerateSharedSecret generates a shared secret based on a secret key and a
// public key using Diffie-Hellman key exchange (ECDH) (RFC 4753).
// RFC5903 Section 9 states we should only return x.
func GenerateSharedSecret(privkey *SecretKey, pubkey *PublicKey) []byte {
	return secp256k1.GenerateSharedSecret(privkey, pubkey)
}
