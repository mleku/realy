#!/usr/bin/env bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
sudo apt -y install build-essential autoconf libtool
cd $SCRIPT_DIR
rm -rf secp256k1
git clone https://github.com/bitcoin-core/secp256k1.git
cd secp256k1
git checkout v0.6.0
git submodule init
git submodule update
./autogen.sh
./configure --enable-module-schnorrsig --enable-module-ecdh --prefix=/usr
make -j1
sudo make install
