# realy.lol

[![Documentation](https://img.shields.io/badge/godoc-documentation-brightgreen.svg)](https://pkg.go.dev/realy.lol)

![realy.png](./realy.png)

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
- custom badger based event store with a garbage collector that prunes off data with least recent
  access
- vanity npub generator that can mine a 5 letter prefix in around 15 minutes on a 6 core Ryzen 5
  processor
- reverse proxy tool with support for Go vanity imports and nip-05 npub DNS verification and own
  TLS certificates

## Building

If you just want to make it run from source, you should check out a tagged version. The commits on these tags will 
explain what state the commit is at. In general, the most stable versions are new minor tags, eg v1.2.0 or v1.23.0, and minor 
patch versions may not be stable and occasionally may not compile (not very often).

## Repository Policy

In general, the main `dev` branch will build, but occasionally may not. It is where new commits are added once they are 
working, mostly, and allows people to easily see ongoing activity. IT IS NOT GUARANTEED TO BE STABLE.

Sometimes there will be a github release version, as well, these will be the most stable version.

Currently this project is in active development and currently v1.2.9 is quite stable but there may be bugs still.

NWC integration is being worked on currently to enable in-app easy subscription management without any extra interface
tooling, just standard nostr client zap and DM functionality, similar to how access control management already is 
configured simply by making a nostr identity and setting its follows to those you want to be able to read and post, and 
mutes to those whose events will never be stored on the relay, no matter if the user publish them to it.

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

### Static build

To produce a static binary, whether you use the CGO secp256k1 or disable CGO as above:

    go build --ldflags '-extldflags "-static"' -o ~/bin/realy ./cmd/realy/.

will place it into your `~/bin/` directory, and it will work on any system of the same architecture with the same glibc major version (has been 2 for a long time).

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