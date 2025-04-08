//go:build cgo

package p256k

import (
	"crypto/rand"
	"unsafe"

	"realy.lol/ec/schnorr"
	"realy.lol/ec/secp256k1"
	"realy.lol/sha256"
)

/*
#cgo LDFLAGS: -lsecp256k1
#include <secp256k1.h>
#include <secp256k1_schnorrsig.h>
#include <secp256k1_extrakeys.h>
*/
import "C"

type (
	Context  = C.secp256k1_context
	Uchar    = C.uchar
	Cint     = C.int
	SecKey   = C.secp256k1_keypair
	PubKey   = C.secp256k1_xonly_pubkey
	ECPubKey = C.secp256k1_pubkey
)

var (
	ctx *Context
)

func CreateContext() *Context {
	return C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN |
		C.SECP256K1_CONTEXT_VERIFY)
}

func GetRandom() (u *Uchar) {
	rnd := make([]byte, 32)
	_, _ = rand.Read(rnd)
	return ToUchar(rnd)
}

func AssertLen(b []byte, length int, name string) (err error) {
	if len(b) != length {
		err = errorf.E("%s should be %d bytes, got %d", name, length, len(b))
	}
	return
}

func RandomizeContext(ctx *C.secp256k1_context) {
	C.secp256k1_context_randomize(ctx, GetRandom())
	return
}

func CreateRandomContext() (c *Context) {
	c = CreateContext()
	RandomizeContext(c)
	return
}

func init() {
	if ctx = CreateContext(); ctx == nil {
		panic("failed to create secp256k1 context")
	}
}

func ToUchar(b []byte) (u *Uchar) { return (*Uchar)(unsafe.Pointer(&b[0])) }

type Sec struct {
	Key SecKey
}

func GenSec() (sec *Sec, err error) {
	if _, _, sec, _, err = Generate(); chk.E(err) {
		return
	}
	return
}

func SecFromBytes(sk []byte) (sec *Sec, err error) {
	sec = new(Sec)
	if C.secp256k1_keypair_create(ctx, &sec.Key, ToUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	return
}

func (s *Sec) Sec() *SecKey { return &s.Key }

func (s *Sec) Pub() (p *Pub, err error) {
	p = new(Pub)
	if C.secp256k1_keypair_xonly_pub(ctx, &p.Key, nil, s.Sec()) != 1 {
		err = errorf.E("pubkey derivation failed")
		return
	}
	return
}

// type PublicKey struct {
// 	Key *C.secp256k1_pubkey
// }
//
// func NewPublicKey() *PublicKey {
// 	return &PublicKey{
// 		Key: &C.secp256k1_pubkey{},
// 	}
// }

type XPublicKey struct {
	Key *C.secp256k1_xonly_pubkey
}

func NewXPublicKey() *XPublicKey {
	return &XPublicKey{
		Key: &C.secp256k1_xonly_pubkey{},
	}
}

// FromSecretBytes parses and processes what should be a secret key. If it is a correct key within the curve order, but
// with a public key having an odd Y coordinate, it returns an error with the fixed key.
func FromSecretBytes(skb []byte) (
	pkb []byte,
	sec *Sec,
	pub *XPublicKey,
	// ecPub *PublicKey,
	err error) {
	xpkb := make([]byte, schnorr.PubKeyBytesLen)
	// clen := C.size_t(secp256k1.PubKeyBytesLenCompressed - 1)
	pkb = make([]byte, schnorr.PubKeyBytesLen)
	var parity Cint
	// ecPub = NewPublicKey()
	pub = NewXPublicKey()
	sec = &Sec{}
	uskb := ToUchar(skb)
	res := C.secp256k1_keypair_create(ctx, &sec.Key, uskb)
	if res != 1 {
		err = errorf.E("failed to create secp256k1 keypair")
		return
	}
	// C.secp256k1_keypair_pub(ctx, ecPub.Key, &sec.Key)
	// C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, ecPub.Key,
	// 	C.SECP256K1_EC_COMPRESSED)
	// if ecpkb[0] != 2 {
	// log.W.ToSliceOfBytes("odd pubkey from %0x -> %0x", skb, ecpkb)
	// 	Negate(skb)
	// 	uskb = ToUchar(skb)
	// 	res = C.secp256k1_keypair_create(ctx, &sec.Key, uskb)
	// 	if res != 1 {
	// 		err = errorf.E("failed to create secp256k1 keypair")
	// 		return
	// 	}
	// 	C.secp256k1_keypair_pub(ctx, ecPub.Key, &sec.Key)
	// 	C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, ecPub.Key, C.SECP256K1_EC_COMPRESSED)
	// 	C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
	// 	err = errors.New("provided secret generates a public key with odd Y coordinate, fixed version returned")
	// }
	C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(xpkb), pub.Key)
	pkb = xpkb
	// log.I.S(sec, pub, skb, pkb)
	return
}

// Generate gathers entropy to generate a full set of bytes and CGO values of it and derived from it to perform
// signature and ECDH operations.
func Generate() (
	skb, pkb []byte,
	sec *Sec,
	pub *XPublicKey,
	err error,
) {
	skb = make([]byte, secp256k1.SecKeyBytesLen)
	pkb = make([]byte, schnorr.PubKeyBytesLen)
	upkb := ToUchar(pkb)
	var parity Cint
	pub = NewXPublicKey()
	sec = &Sec{}
	for {
		if _, err = rand.Read(skb); chk.E(err) {
			return
		}
		uskb := ToUchar(skb)
		if res := C.secp256k1_keypair_create(ctx, &sec.Key, uskb); res != 1 {
			err = errorf.E("failed to create secp256k1 keypair")
			continue
		}
		C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
		C.secp256k1_xonly_pubkey_serialize(ctx, upkb, pub.Key)
		break
	}
	return
}

// Negate inverts a secret key so an odd prefix bit becomes even and vice versa.
func Negate(uskb []byte) { C.secp256k1_ec_seckey_negate(ctx, ToUchar(uskb)) }

type ECPub struct {
	Key ECPubKey
}

// ECPubFromSchnorrBytes converts a BIP-340 public key to its even standard 33 byte encoding.
//
// This function is for the purpose of getting a key to do ECDH from an x-only key.
func ECPubFromSchnorrBytes(xkb []byte) (pub *ECPub, err error) {
	if err = AssertLen(xkb, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = &ECPub{}
	p := append([]byte{0}, xkb...)
	if C.secp256k1_ec_pubkey_parse(ctx, &pub.Key, ToUchar(p),
		secp256k1.PubKeyBytesLenCompressed) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", p)
		log.I.S(pub)
		return
	}
	return
}

// // ECPubFromBytes parses a pubkey from 33 bytes to the bitcoin-core/secp256k1 struct.
// func ECPubFromBytes(pkb []byte) (pub *ECPub, err error) {
// 	if err = AssertLen(pkb, secp256k1.PubKeyBytesLenCompressed, "pubkey"); chk.E(err) {
// 		return
// 	}
// 	pub = &ECPub{}
// 	if C.secp256k1_ec_pubkey_parse(ctx, &pub.Key, ToUchar(pkb),
// 		secp256k1.PubKeyBytesLenCompressed) != 1 {
// 		err = errorf.E("failed to parse pubkey from %0x", pkb)
// 		log.I.S(pub)
// 		return
// 	}
// 	return
// }

// Pub is a schnorr BIP-340 public key.
type Pub struct {
	Key PubKey
}

// PubFromBytes creates a public key from raw bytes.
func PubFromBytes(pk []byte) (pub *Pub, err error) {
	if err = AssertLen(pk, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = new(Pub)
	if C.secp256k1_xonly_pubkey_parse(ctx, &pub.Key, ToUchar(pk)) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pk)
		return
	}
	return
}

// PubB returns the contained public key as bytes.
func (p *Pub) PubB() (b []byte) {
	b = make([]byte, schnorr.PubKeyBytesLen)
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), &p.Key)
	return
}

// Pub returns the public key as a PubKey.
func (p *Pub) Pub() *PubKey { return &p.Key }

// ToBytes returns the contained public key as bytes.
func (p *Pub) ToBytes() (b []byte, err error) {
	b = make([]byte, schnorr.PubKeyBytesLen)
	if C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), p.Pub()) != 1 {
		err = errorf.E("pubkey serialize failed")
		return
	}
	return
}

// Sign a message and return a schnorr BIP-340 64 byte signature.
func Sign(msg *Uchar, sk *SecKey) (sig []byte, err error) {
	sig = make([]byte, schnorr.SignatureSize)
	c := CreateRandomContext()
	if C.secp256k1_schnorrsig_sign32(c, ToUchar(sig), msg, sk,
		GetRandom()) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

// SignFromBytes Signs a message using a provided secret key and message as raw bytes.
func SignFromBytes(msg, sk []byte) (sig []byte, err error) {
	var umsg *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	var sec *Sec
	if sec, err = SecFromBytes(sk); chk.E(err) {
		return
	}
	return Sign(umsg, sec.Sec())
}

// Msg checks that a message hash is correct, and converts it for use with a Signer.
func Msg(b []byte) (id *Uchar, err error) {
	if err = AssertLen(b, sha256.Size, "id"); chk.E(err) {
		return
	}
	id = ToUchar(b)
	return
}

// Sig checks that a signature bytes is correct, and converts it for use with a Signer.
func Sig(b []byte) (sig *Uchar, err error) {
	if err = AssertLen(b, schnorr.SignatureSize, "sig"); chk.E(err) {
		return
	}
	sig = ToUchar(b)
	return
}

// Verify a message signature matches the provided PubKey.
func Verify(msg, sig *Uchar, pk *PubKey) (valid bool) {
	return C.secp256k1_schnorrsig_verify(ctx, sig, msg, 32, pk) == 1
}

// VerifyFromBytes a signature from the raw bytes of the message hash, signature and public key
func VerifyFromBytes(msg, sig, pk []byte) (err error) {
	var umsg, usig *Uchar
	if umsg, err = Msg(msg); chk.E(err) {
		return
	}
	if usig, err = Sig(sig); chk.E(err) {
		return
	}
	var pub *Pub
	if pub, err = PubFromBytes(pk); chk.E(err) {
		return
	}
	valid := Verify(umsg, usig, pub.Pub())
	if !valid {
		err = errorf.E("failed to verify signature")
	}
	return
}

// Zero wipes the memory of a SecKey by overwriting it three times with random data and then
// zeroing it.
func Zero(sk *SecKey) {
	b := (*[96]byte)(unsafe.Pointer(sk))[:96]
	for range 3 {
		rand.Read(b)
		// reverse the order and negate
		lb := len(b)
		l := lb / 2
		for j := range l {
			b[j] = ^b[lb-1-j]
		}
	}
	for i := range b {
		b[i] = 0
	}
}

// Keygen is an implementation of a key miner designed to be used for vanity key generation with X-only BIP-340 keys.
type Keygen struct {
	secBytes, comprPubBytes []byte
	secUchar, cmprPubUchar  *Uchar
	sec                     *Sec
	// ecpub                   *PublicKey
	cmprLen C.size_t
}

// NewKeygen allocates the required buffers for deriving a key. This should only be done once to avoid garbage and make
// the key mining as fast as possible.
//
// This allocates everything and creates proper CGO variables needed for the generate function so they only need to be
// allocated once per thread.
func NewKeygen() (k *Keygen) {
	k = new(Keygen)
	k.cmprLen = C.size_t(secp256k1.PubKeyBytesLenCompressed)
	k.secBytes = make([]byte, secp256k1.SecKeyBytesLen)
	k.comprPubBytes = make([]byte, secp256k1.PubKeyBytesLenCompressed)
	k.secUchar = ToUchar(k.secBytes)
	k.cmprPubUchar = ToUchar(k.comprPubBytes)
	k.sec = &Sec{}
	// k.ecpub = NewPublicKey()
	return
}

// Generate takes a pair of buffers for the secret and ec pubkey bytes and gathers new entropy and returns a valid
// secret key and the compressed pubkey bytes for the partial collision search.
//
// The first byte of pubBytes must be sliced off before deriving the hex/Bech32 forms of the nostr public key.
func (k *Keygen) Generate() (
	sec *Sec,
	pub *XPublicKey,
	pubBytes []byte,
	err error,
) {
	if _, err = rand.Read(k.secBytes); chk.E(err) {
		return
	}
	if res := C.secp256k1_keypair_create(ctx, &k.sec.Key, k.secUchar); res != 1 {
		err = errorf.E("failed to create secp256k1 keypair")
		return
	}
	var parity Cint
	C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
	// C.secp256k1_keypair_pub(ctx, k.ecpub.Key, &k.sec.Key)
	// C.secp256k1_ec_pubkey_serialize(ctx, k.cmprPubUchar, &k.cmprLen, k.ecpub.Key,
	// 	C.SECP256K1_EC_COMPRESSED)
	// pubBytes = k.comprPubBytes
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(pubBytes), pub.Key)
	// pubBytes =
	return
}
