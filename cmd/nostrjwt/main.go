package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/lol"
	"realy.lol/p256k"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

const secEnv = "NOSTR_SECRET_KEY"
const jwtSecEnv = "NOSTR_JWT_SECRET"

var userAgent = fmt.Sprintf("nostrjwt/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	lol.SetLogLevel("trace")
	// log.I.S(os.Args)
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`nostrjwt usage:

nostrjwt gen

	generate a JWT .pem secret and a nostr event for kind %d that creates a binding between the JWT .crt and the nostr npub for authentication, as an alternative to nip-98 on devices that cannot do BIP-340 signatures.

	the .pem data should be stored in %s environment variable for later use

nostrjwt bearer [optional expiry in 0h0m0s format for JWT token]

	using the JWT .pem data, found in the %s environment variable, generate a signed JWT header in standard format as used by curl to add to make GET and POST requests to a nostr HTTP JWT savvy relay to read or publish events.
`, kind.JWTBinding.K, jwtSecEnv, jwtSecEnv)
		os.Exit(0)
	}
	var err error
	var sk []byte
	nsex := os.Getenv(secEnv)
	if len(nsex) == 0 {
		fail("no key found in environment variable %s", secEnv)
	}
	if sk, err = bech32encoding.NsecToBytes([]byte(nsex)); chk.E(err) {
		fail("failed to decode nsec: '%s'", err.Error())
	}
	// log.I.S(nsex, sk)
	sign := &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		fail("failed to init signer: '%s'", err.Error())
	}
	var pk *ecdsa.PrivateKey
	if pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); chk.E(err) {
		fail(err.Error())
	}
	var pkb []byte
	if pkb, err = x509.MarshalPKIXPublicKey(pk.Public()); chk.E(err) {
		fail(err.Error())
	}
	dstp := make([]byte, len(pkb)*8/6+3)
	// log.I.S(pkb)
	base64.URLEncoding.Encode(dstp, pkb)
	// fmt.Printf("%s\n", dstp)
	// log.I.F("%0x %x", pk.X, pk.Y)
	var b []byte
	if b, err = x509.MarshalECPrivateKey(pk); chk.E(err) {
		fail(err.Error())
	}
	dsts := make([]byte, len(b)*8/6+3)
	base64.URLEncoding.Encode(dsts, b)
	fmt.Printf("%s=%s\n", jwtSecEnv, dsts)

	var ev event.T
	ev.Tags = tags.New(tag.New([]byte("J"), dstp, []byte("ES256")))
	ev.CreatedAt = timestamp.Now()
	ev.Kind = kind.JWTBinding
	ev.Sign(sign)
	fmt.Printf("%s\n", ev.SerializeIndented())
}
