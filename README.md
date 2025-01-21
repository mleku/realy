# realy.lol

[![Documentation](https://img.shields.io/badge/godoc-documentation-blue.svg)](https://pkg.go.dev/realy.lol)

[![Support this project](https://img.shields.io/badge/donate-geyser_crowdfunding_project_page-orange.svg)](https://geyser.fund/project/realy)

zap me: ⚡️mleku@getalby.com support through geyser: ⚡️realy@geyser.fund

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
- [libsecp256k1](https://github.com/bitcoin/secp256k1)-enabled signature and signature verification
  (see [here](p256k/README.md))
- a bespoke, mutable byte slice based hash/pubkey/signature encoding in memory
- custom badger based event store with a garbage collector that prunes off data with least recent
  access
- vanity npub generator that can mine a 5 letter prefix in around 15 minutes on a 6 core Ryzen 5
  processor
- reverse proxy tool with support for Go vanity imports
  and [nip-05](https://github.com/nostr-protocol/nips/blob/master/05.md) npub DNS verification and own
  TLS certificates

## Building

If you just want to make it run from source, you should check out a tagged version. The commits on these tags will
explain what state the commit is at. In general, the most stable versions are new minor tags, eg v1.2.0 or v1.23.0, and
minor
patch versions may not be stable and occasionally may not compile (not very often).

The actual executable things are found in the [cmd](cmd/) directory. Currently there is 4 things you can find in there:

- birb - a pure Go GUI for working with NIP-79 (provisional) Nostr Relay Chat protocol built
  using [gio](https://gioui.org) that will eventually work on all platforms, linux X/Wayland, Windows, Mac, iOS, Android
  and WASM browser module
- lerproxy - a very simple to configure reverse proxy that provides SSL/TLS via LetsEncrypt or optionally with your own
  certificates, as well as NIP-05 DNS verification and Go vanity imports
- realy - a nostr relay with a number of unique features, built from forks of the go-nostr library and relayer relay
- vainstr - a vanity key miner to generate bech32 encoded nostr public keys with a chosen text beginning, end or
  anywhere

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

will place it into your `~/bin/` directory, and it will work on any system of the same architecture with the same glibc
major version (has been 2 for a long time).

## Nostr Relay Chat client dependencies

> this chat client is in very early stages, it doesn't do anything yet

You can follow the [directions](https://gioui.org/doc/install) from gioui.org or if you are running an ubuntu/debian
based linux distribution use the script [here](gio/ubuntu_install.sh)

## Configuration

The default will run the relay with default settings, which will not be what you
want.

To see the curent active configuration:

    realy env

This output can be directed to the profile location to make the settings
editable without manually setting them on the commandline:

    realy env > $HOME/.config/realy/.env

You can now edit this file to alter the configuration.

Note the configuration file is a "dotfile" so that if you are tinkering with the
code you can wipe out a broken database with:

    rm -rf $HOME/.config/realy/*

and it leaves the config because this doesn't match a standard wildcard, all the
database files wil be removed, however.

Regarding the configuration system, this is an element of many servers that is
absurdly complex, and for which reason Realy does not use a complicated scheme,
a simple library that allows automatic configuration of a series of options,
added a simple info print:

    realy help

will show you the instructions, and the one simple extension of being able to
use a standard formated .env file to configure all the options for an instance.

## Access Control System

Rather than make a separate interface, and to make personal realy configuration simple, you can simply designate pubkeys
that are listed as comma separated hex public keys in the configuration environment or the `~/.config/realy/.env`
environment configuration file, and it is necessary that the owners publish their follow list, and all npubs named in
the list are whitelisted for read/write access.

If the follow lists of these whitelisted users are also published to the realy, the entirety of these npubs is added to
the list of whitelisted users.

In addition, the npubs in the OWNERS mute lists are removed from the whitelist, but their events may be published to the
relay by any user who is authenticated with an npub in the whitelist that was created as described. Those who are
whitelisted are not prevented from publishing events authored by these users, but the users themselves don't get access.
This is simply the rightful control that a relay owner, who pays the bill, should have over who has access to add data
to the relay, as these can potentially be seeking to attack the owner and flood their database with garbage as a
resource exhaustion attack.

In order to facilitate this functionality, one may also designate a secret key, and if one is designated, this will
enable a spider functionality that will attempt to contact any and all relays found in events on the relay, randomly,
and once it acquires two accessible relays that return nip-11 info and return results from queries of all the currently
whitelisted users on the relay, it goes quiet again until the next hour, where it continues to search and scrape to
update the users latest versions of their follow lists in order to maintain the access control list as current as
possible.

The purpose of this is that it means that any typical nostr client can be used to control the access to a relay, not
just the administrators, but all of the whitelisted users, if they follow an npub, this npub also can automatically
publish to the relay and thus enable an effective use of the nip-65 in/outbox scheme which enables users to rendezvous
for direct messaging at a realy that they are whitelisted as users by the owner.

## Administrative functions

You can export everything in the event store through the default http://localhost:3334 endpoint
like so:

    curl -u username:password http://localhost:3334/export > everything.jsonl

The username and password are configured in the environment variables

    ADMIN_USER=username
    ADMIN_PASSWORD=password

Note that HTTP basic authentication this can only be alphanumeric values, but
make it long and strong because these functions can do bad things to your relay. If these variables are unset (default)
these functions will not be available.

Or just all of the whitelisted users and all events with p tags with them in it:

    curl -u username:password http://localhost:3334/export/users > users.jsonl

Or just one user: (includes also matching p tags)

    curl -u username:password http://localhost:3334/export/4c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f > mleku.jsonl

Or several users with hyphens between the hexadecimal public keys: (ditto above)

    curl -u username:password http://localhost:3334/export/4c800257a588a82849d049817c2bdaad984b25a45ad9f6dad66e47d3b47e3b2f-454bc2771a69e30843d0fccfde6e105ff3edc5c6739983ef61042633e4a9561a > mleku_gojiberra.jsonl

And import also, to put one of these files (also nostrudel and coracle have functions to
export the app database of events in jsonl)

    curl -u username:password -XPOST -T nostrudel.jsonl http://localhost:3334/import

You can also shut down the realy as well:

    curl -u username:password http://localhost:3334/shutdown

Other administrative features will probably be added later, these are just the
essentials.

Other
