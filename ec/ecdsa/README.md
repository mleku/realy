ecdsa
=====

[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/mleku.online/git/ec/secp/ecdsa)

Package ecdsa provides secp256k1-optimized ECDSA signing and verification.

This package provides data structures and functions necessary to produce and
verify deterministic canonical signatures in accordance with RFC6979 and
BIP0062, optimized specifically for the secp256k1 curve using the Elliptic Curve
Digital Signature Algorithm (ECDSA), as defined in FIPS 186-3.  See
https://www.secg.org/sec2-v2.pdf (also found here at [sec2-v2.pdf](../sec2-v2.pdf)) for details on the secp256k1 standard.

It also provides functions to parse and serialize the ECDSA signatures with the
more strict Distinguished Encoding Rules (DER) of ISO/IEC 8825-1 and some
additional restrictions specific to secp256k1.

In addition, it supports a custom "compact" signature format which allows
efficient recovery of the public key from a given valid signature and message
hash combination.

A comprehensive suite of tests is provided to ensure proper functionality.

## License

Package ecdsa is licensed under the [copyfree](http://copyfree.org) ISC License.
