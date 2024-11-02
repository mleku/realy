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

## CGO and secp256k1 signatures library

By default, Go will usually be configured with `CGO_ENABLED=1`. This selects the use of the 
C library from bitcoin core, which does signatures and verifications much faster (4x and better)
but complicates the build process as you have to install the library beforehand. There is
instructions in [p256k/README.md](p256k/README.md) for doing this.

In order to disable the use of this, you must set the environment variable `CGO_ENABLED=0` and
it the Go compiler will automatically revert to using the btcec based secp256k1 signatures 
library.

    export CGO_ENABLED=0
    cd cmd/realy
    go build .

This will build the binary and place it in cmd/realy and then you can move it where you like.

## Export and Import functions

You can export everything in the event store through the default http://localhost:3337 endpoint
like so:

    curl http://localhost:3337/export > everything.jsonl

or just all of the whitelisted users and all events with p tags with them in it:

    curl http://localhost:3337/export/users > users.jsonl

or just one user: (includes also matching p tags)

    curl http://localhost:3337/export/4c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f > mleku.jsonl

or several users with hyphens between the hexadecimal public keys: (ditto above)

    curl http://localhost:3337/export/4c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f-454bc2771a69e30843d0fccfde6e105ff3edc5c6739983ef61042633e4a9561a > mleku_gojiberra.jsonl


and import also, to put one of these files (also nostrudel and coracle have functions to 
export the app database of events in jsonl)

    curl -XPOST -T nostrudel.jsonl http://localhost:3337/import

> todo: more documentation coming