# p256k1

This is a library that uses the `bitcoin-core` optimized secp256k1 elliptic
curve signatures library for `nostr` schnorr signatures.

By default it uses the secp256k1 EC library from
[btcsuite](https://github.com/btcsuite/btcd)/[decred](https://github.com/decred/dcrd/tree/master/dcrec), 
the decred is actually where the schnorr signatures are (ikr?) - this repo 
uses my fork of this mess of shitcoinery and bad, slow Go code is cleaned up 
and unified in [realy.lol/ec](https://realy.lol/ec) and 
includes the bech32 and base58check libraries. And the messy precomputed 
values are upgraded to use the modern "embed" enabling a faster app startup 
for initialising this array (at the cost of a little more binary size).

For ubuntu, you need these

    sudo apt -y install build-essential autoconf libtool  

The directory `pkg/libsecp256k1/secp256k1` needs to be initialized and built
and installed, like so:

```bash
cd p256k
git clone https://github.com/bitcoin-core/secp256k1.git
cd secp256k1
git submodule init
git submodule update
```

Then to build, you can refer to the [instructions](./secp256k1/README.md) or
just use the default autotools:

```bash
./autogen.sh
./configure --enable-module-schnorrsig --prefix=/usr
make
sudo make install
```

On WSL2 you may have to attend to various things to make this work, setting up your basic locale (uncomment one or more in `/etc/locale.gen`, and run `locale-gen`), installing the basic build tools (build-essential or base-devel) and of course git, curl, wget, libtool and autoconf.