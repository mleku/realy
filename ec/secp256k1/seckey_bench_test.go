// Copyright (c) 2022 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package secp256k1

import (
	"testing"
)

// BenchmarkSecretKeyGenerate benchmarks generating new cryptographically
// secure secret keys.
func BenchmarkSecretKeyGenerate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateSecretKey()
		if err != nil {
			b.Fatal(err)
		}
	}
}
