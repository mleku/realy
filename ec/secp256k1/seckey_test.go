// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2023 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package secp256k1

import (
	"bytes"
	"crypto/rand"
	"errors"
	"math/big"
	"testing"
)

// TestGenerateSecretKey ensures the key generation works as expected.
func TestGenerateSecretKey(t *testing.T) {
	sec, err := GenerateSecretKey()
	if err != nil {
		t.Errorf("failed to generate secret key: %s", err)
		return
	}
	pub := sec.PubKey()
	if !isOnCurve(&pub.x, &pub.y) {
		t.Error("public key is not on the curve")
	}
}

// TestGenerateSecretKeyFromRand ensures generating a secret key from a random
// entropy source works as expected.
func TestGenerateSecretKeyFromRand(t *testing.T) {
	sec, err := GenerateSecretKeyFromRand(rand.Reader)
	if err != nil {
		t.Errorf("failed to generate secret key: %s", err)
		return
	}
	pub := sec.PubKey()
	if !isOnCurve(&pub.x, &pub.y) {
		t.Error("public key is not on the curve")
	}
}

// mockSecretKeyReaderFunc is an adapter to allow the use of an ordinary
// function as an io.Reader.
type mockSecretKeyReaderFunc func([]byte) (int, error)

// Read calls the function with the provided parameter and returns the result.
func (f mockSecretKeyReaderFunc) Read(p []byte) (int, error) {
	return f(p)
}

// TestGenerateSecretKeyCorners ensures random values that secret key
// generation correctly handles entropy values that are invalid for use as
// secret keys by creating a fake source of randomness to inject known bad
// values.
func TestGenerateSecretKeyCorners(t *testing.T) {
	// Create a mock reader that returns the following sequence of values:
	// 1st invocation: 0
	// 2nd invocation: The curve order
	// 3rd invocation: The curve order + 1
	// 4th invocation: 1 (32-byte big endian)
	oneModN := hexToModNScalar("01")
	var numReads int
	mockReader := mockSecretKeyReaderFunc(func(p []byte) (int, error) {
		numReads++
		switch numReads {
		case 1:
			return copy(p, bytes.Repeat([]byte{0x00}, len(p))), nil
		case 2:
			return copy(p, curveParams.N.Bytes()), nil
		case 3:
			nPlusOne := new(big.Int).Add(curveParams.N, big.NewInt(1))
			return copy(p, nPlusOne.Bytes()), nil
		}
		oneModNBytes := oneModN.Bytes()
		return copy(p, oneModNBytes[:]), nil
	})
	// Generate a secret key using the mock reader and ensure the resulting key
	// is the expected one.  It should be the value "1" since the other values
	// the sequence produces are invalid and thus should be rejected.
	sec, err := GenerateSecretKeyFromRand(mockReader)
	if err != nil {
		t.Errorf("failed to generate secret key: %s", err)
		return
	}
	if !sec.Key.Equals(oneModN) {
		t.Fatalf("unexpected secret key -- got: %x, want %x", sec.Serialize(),
			oneModN.Bytes())
	}
}

// TestGenerateSecretKeyError ensures the secret key generation properly
// handles errors when attempting to read from the source of randomness.
func TestGenerateSecretKeyError(t *testing.T) {
	// Create a mock reader that returns an error.
	errDisabled := errors.New("disabled")
	mockReader := mockSecretKeyReaderFunc(func(p []byte) (int, error) {
		return 0, errDisabled
	})
	// Generate a secret key using the mock reader and ensure the expected
	// error is returned.
	_, err := GenerateSecretKeyFromRand(mockReader)
	if !errors.Is(err, errDisabled) {
		t.Fatalf("mismatched err -- got %v, want %v", err, errDisabled)
		return
	}
}

// TestSecKeys ensures a secret key created from bytes produces both the
// correct associated public key as well serializes back to the original bytes.
func TestSecKeys(t *testing.T) {
	tests := []struct {
		name string
		sec  string // hex encoded secret key to test
		pub  string // expected hex encoded serialized compressed public key
	}{{
		name: "random secret key 1",
		sec:  "eaf02ca348c524e6392655ba4d29603cd1a7347d9d65cfe93ce1ebffdca22694",
		pub:  "025ceeba2ab4a635df2c0301a3d773da06ac5a18a7c3e0d09a795d7e57d233edf1",
	}, {
		name: "random secret key 2",
		sec:  "24b860d0651db83feba821e7a94ba8b87162665509cefef0cbde6a8fbbedfe7c",
		pub:  "032a6e51bf218085647d330eac2fafaeee07617a777ad9e8e7141b4cdae92cb637",
	}}

	for _, test := range tests {
		// Parse test data.
		secKeyBytes := hexToBytes(test.sec)
		wantPubKeyBytes := hexToBytes(test.pub)

		sec := SecKeyFromBytes(secKeyBytes)
		pub := sec.PubKey()

		serializedPubKey := pub.SerializeCompressed()
		if !bytes.Equal(serializedPubKey, wantPubKeyBytes) {
			t.Errorf("%s unexpected serialized public key - got: %x, want: %x",
				test.name, serializedPubKey, wantPubKeyBytes)
		}

		serializedSecKey := sec.Serialize()
		if !bytes.Equal(serializedSecKey, secKeyBytes) {
			t.Errorf("%s unexpected serialized secret key - got: %x, want: %x",
				test.name, serializedSecKey, secKeyBytes)
		}
	}
}

// TestSecretKeyZero ensures that zeroing a secret key clears the memory
// associated with it.
func TestSecretKeyZero(t *testing.T) {
	// Create a new secret key and zero the initial key material that is now
	// copied into the secret key.
	key := new(ModNScalar).SetHex("eaf02ca348c524e6392655ba4d29603cd1a7347d9d65cfe93ce1ebffdca22694")
	secKey := NewSecretKey(key)
	key.Zero()
	// Ensure the secret key is non zero.
	if secKey.Key.IsZero() {
		t.Fatal("secret key is zero when it should be non zero")
	}
	// Zero the secret key and ensure it was properly zeroed.
	secKey.Zero()
	if !secKey.Key.IsZero() {
		t.Fatal("secret key is non zero when it should be zero")
	}
}
