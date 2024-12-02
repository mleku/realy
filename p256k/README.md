# p256k1

This is a library that uses the `bitcoin-core` optimized secp256k1 elliptic
curve signatures library for `nostr` schnorr signatures.

If you need to build it without `libsecp256k1` C library, you must disable cgo:

    export CGO_ENABLED='0'

This enables the fallback `btcec` pure Go library to be used in its place. This
CGO setting is not default for Go, so it must be set in order to disable this.

The standard `libsecp256k1-0` and `libsecp256k1-dev` available through the
ubuntu dpkg repositories do not include support for the BIP-340 schnorr
signatures or the ECDH X-only shared secret generation algorithm, so you must
follow the following instructions to get the benefits of using this library. It
is 4x faster at signing and generating shared secrets so it is a must if your
intention is to use it for high throughput systems like a network transport.

The easy way to install it, if you have ubuntu/debian, is the script
[../ubuntu_install_libsecp256k1.sh](../ubuntu_install_libsecp256k1.sh), it
handles the dependencies and runs the build all in one step for you. Note that it 

For ubuntu, you need these:

    sudo apt -y install build-essential autoconf libtool  

For other linux distributions, the process is the same but the dependencies are
likely different. The main thing is it requires make, gcc/++, autoconf and
libtool to run. The most important thing to point out is that you must enable
the schnorr signatures feature, and ECDH.

The directory `p256k/secp256k1` needs to be initialized, built and installed,
like so:

```bash
cd secp256k1
git submodule init
git submodule update
```

Then to build, you can refer to the [instructions](./secp256k1/README.md) or
just use the default autotools:

```bash
./autogen.sh
./configure --enable-module-schnorrsig --enable-module-ecdh --prefix=/usr
make
sudo make install
```

On WSL2 you may have to attend to various things to make this work, setting up
your basic locale (uncomment one or more in `/etc/locale.gen`, and run
`locale-gen`), installing the basic build tools (build-essential or base-devel)
and of course git, curl, wget, libtool and
autoconf.

## ECDH

TODO: Currently the use of the libsecp256k1 library for ECDH, used in nip-04 and
nip-44 encryption is not enabled, because the default version uses the Y
coordinate and this is incorrect for nostr. It will be enabled soon... for now
it is done with the `btcec` fallback version. This is slower, however previous 
tests have shown that this ECDH library is fast enough to enable 8mb/s 
throughput per CPU thread when used to generate a distinct secret for TCP 
packets. The C library will likely raise this to 20mb/s or more.