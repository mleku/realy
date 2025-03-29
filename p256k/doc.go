// Package p256k is a signer interface that (by default) uses the
// bitcoin/libsecp256k1 library for fast signature creation and verification of
// the BIP-340 nostr X-only signatures and public keys, and ECDH.
//
// Currently the ECDH is only implemented with the btcec library.
package p256k
