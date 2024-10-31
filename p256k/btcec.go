//go:build !cgo

package p256k

import (
	"realy.lol/p256k/btcec"
)

// BTCECSigner is always available but enabling it disables the use of github.com/bitcoin-core/secp256k1 CGO signature
// implementation and points it at the btec version.

type Signer = btcec.Signer
type Keygen = btcec.Keygen

func NewKeygen() (k *Keygen) { return new(Keygen) }
