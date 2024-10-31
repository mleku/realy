//go:build cgo

package p256k

import "C"
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

func AssertLen(b B, length int, name string) (err E) {
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

func ToUchar(b B) (u *Uchar) { return (*Uchar)(unsafe.Pointer(&b[0])) }

type Sec struct {
	Key SecKey
}

func GenSec() (sec *Sec, err E) {
	if _, _, sec, _, _, err = Generate(); chk.E(err) {
		return
	}
	return
}

func SecFromBytes(sk B) (sec *Sec, err E) {
	sec = new(Sec)
	if C.secp256k1_keypair_create(ctx, &sec.Key, ToUchar(sk)) != 1 {
		err = errorf.E("failed to parse private key")
		return
	}
	return
}

func (s *Sec) Sec() *SecKey { return &s.Key }

func (s *Sec) Pub() (p *Pub, err E) {
	p = new(Pub)
	if C.secp256k1_keypair_xonly_pub(ctx, &p.Key, nil, s.Sec()) != 1 {
		err = errorf.E("pubkey derivation failed")
		return
	}
	return
}

type PublicKey struct {
	Key *C.secp256k1_pubkey
}

func NewPublicKey() *PublicKey {
	return &PublicKey{
		Key: &C.secp256k1_pubkey{},
	}
}

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
func FromSecretBytes(skb B) (pkb B, sec *Sec, pub *XPublicKey, ecPub *PublicKey, err error) {
	ecpkb := make(B, secp256k1.PubKeyBytesLenCompressed)
	clen := C.size_t(secp256k1.PubKeyBytesLenCompressed)
	pkb = make(B, secp256k1.PubKeyBytesLenCompressed)
	var parity Cint
	ecPub = NewPublicKey()
	pub = NewXPublicKey()
	sec = &Sec{}
	uskb := ToUchar(skb)
	res := C.secp256k1_keypair_create(ctx, &sec.Key, uskb)
	if res != 1 {
		err = errorf.E("failed to create secp256k1 keypair")
		return
	}
	C.secp256k1_keypair_pub(ctx, ecPub.Key, &sec.Key)
	C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, ecPub.Key,
		C.SECP256K1_EC_COMPRESSED)
	// if ecpkb[0] != 2 {
	// log.W.F("odd pubkey from %0x -> %0x", skb, ecpkb)
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
	// log.I.S(sec, ecPub, pub)
	pkb = ecpkb
	return
}

// Generate gathers entropy to generate a full set of bytes and CGO values of it and derived from it to perform
// signature and ECDH operations.
//
// Note that the pubkey bytes are the 33 byte form with the sign prefix, slice it off for X-only use.
func Generate() (skb, pkb B, sec *Sec, pub *XPublicKey, ecpub *PublicKey, err E) {
	skb = make(B, secp256k1.SecKeyBytesLen)
	ecpkb := make(B, secp256k1.PubKeyBytesLenCompressed)
	clen := C.size_t(secp256k1.PubKeyBytesLenCompressed)
	pkb = make(B, secp256k1.PubKeyBytesLenCompressed)
	var parity Cint
	ecpub = NewPublicKey()
	pub = NewXPublicKey()
	sec = &Sec{}
	for {
		if _, err = rand.Read(skb); chk.E(err) {
			return
		}
		uskb := ToUchar(skb)
		if res := C.secp256k1_keypair_create(ctx, &sec.Key, uskb); res != 1 {
			err = errorf.E("failed to create secp256k1 keypair")
			return
		}
		C.secp256k1_keypair_pub(ctx, ecpub.Key, &sec.Key)
		C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, ecpub.Key,
			C.SECP256K1_EC_COMPRESSED)
		// negate key if it generates an odd Y compressed public key per BIP-340 (or at least so ecdsa can assume - and
		// maybe it's a tiny bit faster to verify? idk, bip-340 says "implicit 0x02 key" so whatever let's make em)
		if ecpkb[0] == 2 {
			C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
			break
		} else {
			Negate(skb)
			C.secp256k1_keypair_pub(ctx, ecpub.Key, &sec.Key)
			C.secp256k1_ec_pubkey_serialize(ctx, ToUchar(ecpkb), &clen, ecpub.Key,
				C.SECP256K1_EC_COMPRESSED)
			C.secp256k1_keypair_xonly_pub(ctx, pub.Key, &parity, &sec.Key)
			break
		}
	}
	pkb = ecpkb
	return
}

func Negate(uskb B) { C.secp256k1_ec_seckey_negate(ctx, ToUchar(uskb)) }

type ECPub struct {
	Key ECPubKey
}

// ECPubFromSchnorrBytes converts a BIP-340 public key to its even standard 33 byte encoding.
//
// This function is for the purpose of getting a key to do ECDH from an x-only key.
func ECPubFromSchnorrBytes(xkb B) (pub *ECPub, err E) {
	if err = AssertLen(xkb, schnorr.PubKeyBytesLen, "pubkey"); chk.E(err) {
		return
	}
	pub = &ECPub{}
	p := append(B{0x02}, xkb...)
	if C.secp256k1_ec_pubkey_parse(ctx, &pub.Key, ToUchar(p),
		secp256k1.PubKeyBytesLenCompressed) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", p)
		log.I.S(pub)
		return
	}
	return
}

// ECPubFromBytes parses a pubkey from 33 bytes to the bitcoin-core/secp256k1 struct.
func ECPubFromBytes(pkb B) (pub *ECPub, err E) {
	if err = AssertLen(pkb, secp256k1.PubKeyBytesLenCompressed, "pubkey"); chk.E(err) {
		return
	}
	pub = &ECPub{}
	if C.secp256k1_ec_pubkey_parse(ctx, &pub.Key, ToUchar(pkb),
		secp256k1.PubKeyBytesLenCompressed) != 1 {
		err = errorf.E("failed to parse pubkey from %0x", pkb)
		log.I.S(pub)
		return
	}
	return
}

type Pub struct {
	Key PubKey
}

func PubFromBytes(pk B) (pub *Pub, err E) {
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

func (p *Pub) PubB() (b B) {
	b = make(B, schnorr.PubKeyBytesLen)
	C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), &p.Key)
	return
}

func (p *Pub) Pub() *PubKey { return &p.Key }

func (p *Pub) ToBytes() (b B, err E) {
	b = make(B, schnorr.PubKeyBytesLen)
	if C.secp256k1_xonly_pubkey_serialize(ctx, ToUchar(b), p.Pub()) != 1 {
		err = errorf.E("pubkey serialize failed")
		return
	}
	return
}

func Sign(msg *Uchar, sk *SecKey) (sig B, err E) {
	sig = make(B, schnorr.SignatureSize)
	c := CreateRandomContext()
	if C.secp256k1_schnorrsig_sign32(c, ToUchar(sig), msg, sk,
		GetRandom()) != 1 {
		err = errorf.E("failed to sign message")
		return
	}
	return
}

func SignFromBytes(msg, sk B) (sig B, err E) {
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

func Msg(b B) (id *Uchar, err E) {
	if err = AssertLen(b, sha256.Size, "id"); chk.E(err) {
		return
	}
	id = ToUchar(b)
	return
}

func Sig(b B) (sig *Uchar, err E) {
	if err = AssertLen(b, schnorr.SignatureSize, "sig"); chk.E(err) {
		return
	}
	sig = ToUchar(b)
	return
}

func Verify(msg, sig *Uchar, pk *PubKey) (valid bool) {
	return C.secp256k1_schnorrsig_verify(ctx, sig, msg, 32, pk) == 1
}

func VerifyFromBytes(msg, sig, pk B) (err error) {
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

func Zero(sk *SecKey) {
	b := (*[96]byte)(unsafe.Pointer(sk))[:96]
	for i := range b {
		b[i] = 0
	}
}

// Keygen is an implementation of a key miner designed to be used for vanity key generation with X-only BIP-340 keys.
type Keygen struct {
	secBytes, comprPubBytes B
	secUchar, cmprPubUchar  *Uchar
	sec                     *Sec
	ecpub                   *PublicKey
	cmprLen                 C.size_t
}

// NewKeygen allocates the required buffers for deriving a key. This should only be done once to avoid garbage and make
// the key mining as fast as possible.
//
// This allocates everything and creates proper CGO variables needed for the generate function so they only need to be
// allocated once per thread.
func NewKeygen() (k *Keygen) {
	k = new(Keygen)
	k.cmprLen = C.size_t(secp256k1.PubKeyBytesLenCompressed)
	k.secBytes = make(B, secp256k1.SecKeyBytesLen)
	k.comprPubBytes = make(B, secp256k1.PubKeyBytesLenCompressed)
	k.secUchar = ToUchar(k.secBytes)
	k.cmprPubUchar = ToUchar(k.comprPubBytes)
	k.sec = &Sec{}
	k.ecpub = NewPublicKey()
	return
}

// Generate takes a pair of buffers for the secret and ec pubkey bytes and gathers new entropy and returns a valid
// secret key and the compressed pubkey bytes for the partial collision search.
//
// The first byte of pubBytes must be sliced off before deriving the hex/Bech32 forms of the nostr public key.
func (k *Keygen) Generate() (pubBytes B, err E) {
	if _, err = rand.Read(k.secBytes); chk.E(err) {
		return
	}
	if res := C.secp256k1_keypair_create(ctx, &k.sec.Key, k.secUchar); res != 1 {
		err = errorf.E("failed to create secp256k1 keypair")
		return
	}
	C.secp256k1_keypair_pub(ctx, k.ecpub.Key, &k.sec.Key)
	C.secp256k1_ec_pubkey_serialize(ctx, k.cmprPubUchar, &k.cmprLen, k.ecpub.Key,
		C.SECP256K1_EC_COMPRESSED)
	pubBytes = k.comprPubBytes
	return
}

// Negate should be called when the pubkey's X coordinate is a match but the prefix is a 3. The X coordinate will not
// change but this ensures that when the X-only key has a 2 prefix added for ECDH and other purposes that it works
// correctly. This can be done after a match is found as it does not impact anything except the first byte.
func (k *Keygen) Negate() { C.secp256k1_ec_seckey_negate(ctx, k.secUchar) }

func (k *Keygen) KeyPairBytes() (secBytes, cmprPubBytes B) {
	C.secp256k1_ec_pubkey_serialize(ctx, k.cmprPubUchar, &k.cmprLen, k.ecpub.Key,
		C.SECP256K1_EC_COMPRESSED)
	return k.secBytes, k.comprPubBytes[1:]
}
