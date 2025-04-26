package encryption

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/hkdf"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/sha256"
)

const (
	version          byte = 2
	MinPlaintextSize      = 0x0001 // 1b msg => padded to 32b
	MaxPlaintextSize      = 0xffff // 65535 (64kb-1) => padded to 64kb
)

type Opts struct {
	err   error
	nonce []byte
}

// Deprecated: use WithCustomNonce instead of WithCustomSalt, so the naming is less confusing
var WithCustomSalt = WithCustomNonce

// WithCustomNonce enables using a custom nonce (salt) instead of using the
// system crypto/rand entropy source.
func WithCustomNonce(salt []byte) func(opts *Opts) {
	return func(opts *Opts) {
		if len(salt) != 32 {
			opts.err = errorf.E("salt must be 32 bytes, got %d", len(salt))
		}
		opts.nonce = salt
	}
}

// Encrypt data using a provided symmetric conversation key using NIP-44
// encryption (chacha20 cipher stream and sha256 HMAC).
func Encrypt(plaintext string, conversationKey []byte,
	applyOptions ...func(opts *Opts)) (cipherString string,
	err error) {

	var o Opts
	for _, apply := range applyOptions {
		apply(&o)
	}
	if chk.E(o.err) {
		err = o.err
		return
	}
	if o.nonce == nil {
		o.nonce = make([]byte, 32)
		if _, err = rand.Read(o.nonce); chk.E(err) {
			return
		}
	}
	var enc, cc20nonce, auth []byte
	if enc, cc20nonce, auth, err = getKeys(conversationKey, o.nonce); chk.E(err) {
		return
	}
	plain := []byte(plaintext)
	size := len(plain)
	if size < MinPlaintextSize || size > MaxPlaintextSize {
		err = errorf.E("plaintext should be between 1b and 64kB")
		return
	}
	padding := CalcPadding(size)
	padded := make([]byte, 2+padding)
	binary.BigEndian.PutUint16(padded, uint16(size))
	copy(padded[2:], plain)
	var cipher []byte
	if cipher, err = encrypt(enc, cc20nonce, padded); chk.E(err) {
		return
	}
	var mac []byte
	if mac, err = sha256Hmac(auth, cipher, o.nonce); chk.E(err) {
		return
	}
	ct := make([]byte, 0, 1+32+len(cipher)+32)
	ct = append(ct, version)
	ct = append(ct, o.nonce...)
	ct = append(ct, cipher...)
	ct = append(ct, mac...)
	cipherString = base64.StdEncoding.EncodeToString(ct)
	return
}

// Decrypt data that has been encoded using a provided symmetric conversation
// key using NIP-44 encryption (chacha20 cipher stream and sha256 HMAC).
func Decrypt(b64ciphertextWrapped string, conversationKey []byte) (plaintext string,
	err error) {
	cLen := len(b64ciphertextWrapped)
	if cLen < 132 || cLen > 87472 {
		err = errorf.E("invalid payload length: %d", cLen)
		return
	}
	if b64ciphertextWrapped[:1] == "#" {
		err = errorf.E("unknown version")
		return
	}
	var decoded []byte
	if decoded, err = base64.StdEncoding.DecodeString(b64ciphertextWrapped); chk.E(err) {
		return
	}
	if decoded[0] != version {
		err = errorf.E("unknown version %d", decoded[0])
		return
	}
	dLen := len(decoded)
	if dLen < 99 || dLen > 65603 {
		err = errorf.E("invalid data length: %d", dLen)
		return
	}
	nonce, ciphertext, givenMac := decoded[1:33], decoded[33:dLen-32], decoded[dLen-32:]
	var enc, cc20nonce, auth []byte
	if enc, cc20nonce, auth, err = getKeys(conversationKey, nonce); chk.E(err) {
		return
	}
	var expectedMac []byte
	if expectedMac, err = sha256Hmac(auth, ciphertext, nonce); chk.E(err) {
		return
	}
	if !bytes.Equal(givenMac, expectedMac) {
		err = errorf.E("invalid hmac")
		return
	}
	var padded []byte
	if padded, err = encrypt(enc, cc20nonce, ciphertext); chk.E(err) {
		return
	}
	unpaddedLen := binary.BigEndian.Uint16(padded[0:2])
	if unpaddedLen < uint16(MinPlaintextSize) || unpaddedLen > uint16(MaxPlaintextSize) ||
		len(padded) != 2+CalcPadding(int(unpaddedLen)) {
		err = errorf.E("invalid padding")
		return
	}
	unpadded := padded[2:][:unpaddedLen]
	if len(unpadded) == 0 || len(unpadded) != int(unpaddedLen) {
		err = errorf.E("invalid padding")
		return
	}
	plaintext = string(unpadded)
	return
}

// GenerateConversationKey performs an ECDH key generation hashed with the nip-44-v2 using hkdf.
func GenerateConversationKey(pkh, skh string) (ck []byte, err error) {
	if skh >= "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141" ||
		skh == "0000000000000000000000000000000000000000000000000000000000000000" {
		err = errorf.E("invalid private key: x coordinate %s is not on the secp256k1 curve",
			skh)
		return
	}
	var shared []byte
	if shared, err = ComputeSharedSecret(pkh, skh); chk.E(err) {
		return
	}
	ck = hkdf.Extract(sha256.New, shared, []byte("nip44-v2"))
	return
}

func encrypt(key, nonce, message []byte) (dst []byte, err error) {
	var cipher *chacha20.Cipher
	if cipher, err = chacha20.NewUnauthenticatedCipher(key, nonce); chk.E(err) {
		return
	}
	dst = make([]byte, len(message))
	cipher.XORKeyStream(dst, message)
	return
}

func sha256Hmac(key, ciphertext, nonce []byte) (h []byte, err error) {
	if len(nonce) != sha256.Size {
		err = errorf.E("nonce aad must be 32 bytes")
		return
	}
	hm := hmac.New(sha256.New, key)
	hm.Write(nonce)
	hm.Write(ciphertext)
	h = hm.Sum(nil)
	return
}

func getKeys(conversationKey, nonce []byte) (enc, cc20nonce, auth []byte, err error) {
	if len(conversationKey) != 32 {
		err = errorf.E("conversation key must be 32 bytes")
		return
	}
	if len(nonce) != 32 {
		err = errorf.E("nonce must be 32 bytes")
		return
	}
	r := hkdf.Expand(sha256.New, conversationKey, nonce)
	enc = make([]byte, 32)
	if _, err = io.ReadFull(r, enc); chk.E(err) {
		return
	}
	cc20nonce = make([]byte, 12)
	if _, err = io.ReadFull(r, cc20nonce); chk.E(err) {
		return
	}
	auth = make([]byte, 32)
	if _, err = io.ReadFull(r, auth); chk.E(err) {
		return
	}
	return
}

// CalcPadding creates padding for the message payload that is precisely a power
// of two in order to reduce the chances of plaintext attack. This is plainly
// retarded because it could blow out the message size a lot when just a random few
// dozen bytes and a length prefix would achieve the same result.
func CalcPadding(sLen int) (l int) {
	if sLen <= 32 {
		return 32
	}
	nextPower := 1 << int(math.Floor(math.Log2(float64(sLen-1)))+1)
	chunk := int(math.Max(32, float64(nextPower/8)))
	l = chunk * int(math.Floor(float64((sLen-1)/chunk))+1)
	return
}
