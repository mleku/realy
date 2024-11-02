# realy.lol

nostr relay built from a heavily modified fork
of [nbd-wtf/go-nostr](https://github.com/nbd-wtf/go-nostr)
and [fiatjaf/relayer](https://github.com/fiatjaf/relayer) aimed at maximum performance,
simplicity and memory efficiency.

includes:

- a lot of other bits and pieces accumulated from nearly 8 years of working with Go, logging and
  run control, user data directories (windows, mac, linux, android)
- a cleaned up and unified fork of the btcd/dcred BIP-340 signatures, including the use of
  bitcoin core's BIP-340 implementation (more than 4x faster than btcd)
- AVX/AVX2 optimized SHA256 and SIMD hex encoder
- a bespoke, mutable byte slice based hash/pubkey/signature encoding in memory and the fastest
  nostr binary codec that exists
- custom badger based database with a garbage collector that prunes off data with least recent
  access
- vanity npub generator that can mine a 5 letter prefix in around 15 minutes on a 6 core Ryzen 5
  processor
- reverse proxy tool with support for Go vanity imports and nip-05 npub DNS verification and own
  TLS certificates
