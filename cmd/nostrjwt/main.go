package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/event"
	"realy.lol/httpauth"
	"realy.lol/kind"
	"realy.lol/lol"
	"realy.lol/p256k"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

const (
	secEnv    = "NOSTR_SECRET_KEY"
	jwtSecEnv = "NOSTR_JWT_SECRET"
)

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

	generate a JWT secret and a nostr event for kind %d that creates a binding between the JWT pubkey and the nostr npub for authentication, as an alternative to nip-98 on devices that cannot do BIP-340 signatures.

	the secret key data should be stored in %s environment variable for later use

	the pubkey in the nostr event is required for generating a token

nostrjwt bearer <request URL> <nostr pubkey> [<optional expiry in 0h0m0s format for JWT token>]

	request URL must match the one that will be in the HTTP Request this bearer token must refer to

	nostr pubkey must be registered with the relay as associated with the JWT secret signing the token

	using the JWT secret, found in the %s environment variable, generate a signed JWT header in standard format as used by curl to add to make GET and POST requests to a nostr HTTP JWT savvy relay to read or publish events.

	expiry sets an amount of time after the current moment that the token will expire
`, kind.JWTBinding.K, jwtSecEnv, jwtSecEnv)
		os.Exit(0)
	}
	var err error
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "gen":
			// check environment for secret key
			var skb []byte
			nsex := os.Getenv(secEnv)
			if len(nsex) == 0 {
				fail("no key found in environment variable %s", secEnv)
			}
			if skb, err = bech32encoding.NsecToBytes([]byte(nsex)); chk.E(err) {
				fail("failed to decode nsec: '%s'", err.Error())
			}
			sign := &p256k.Signer{}
			if err = sign.InitSec(skb); chk.E(err) {
				fail("failed to init signer: '%s'", err.Error())
			}
			// generate a new JWT key pair
			var x509sec, x509pub, pemSec, pemPub []byte
			if x509sec, x509pub, pemSec, pemPub, err = httpauth.GenerateJWTKeys(); chk.E(err) {
				fail(err.Error())
			}
			fmt.Printf("%s\n%s\n", pemSec, pemPub)
			fmt.Printf("%s=%s\n\n", jwtSecEnv, x509sec)

			var ev event.T
			ev.Tags = tags.New(tag.New([]byte("J"), x509pub, []byte("ES256")))
			ev.CreatedAt = timestamp.Now()
			ev.Kind = kind.JWTBinding
			if err = ev.Sign(sign); chk.E(err) {
				fail(err.Error())
			}
			fmt.Printf("%s\n", ev.Serialize())

		case "bearer":
			// check args
			if len(os.Args) < 4 {
				fail("missing required positional arguments, got '%s' require 'bearer <request URL> <nostr pubkey>'",
					os.Args[1:])
			}
			// jwt secret key must be found in NOSTR_JWT_SECRET
			var jskb []byte
			jwtSec := os.Getenv(jwtSecEnv)
			if len(jwtSec) == 0 {
				fail("no key found in environment variable %s", jwtSecEnv)
			}
			if jskb, err = base64.URLEncoding.DecodeString(jwtSec); chk.E(err) {
				fail(err.Error())
			}
			var sec *ecdsa.PrivateKey
			if sec, err = x509.ParseECPrivateKey(jskb); chk.E(err) {
				fail(err.Error())
			}
			_ = sec // todo
			var tok []byte
			// generate claim
			if len(os.Args) < 5 {
				tok, err = httpauth.GenerateJWTtoken(os.Args[2], os.Args[3])
			} else if len(os.Args) > 4 {
				tok, err = httpauth.GenerateJWTtoken(os.Args[2], os.Args[3], os.Args[4])
			}

			// fmt.Printf("%s\n", tok)

			var claims jwt.MapClaims
			if err = json.Unmarshal(tok, &claims); chk.E(err) {
				fail(err.Error())
			}
			// log.I.S(claims)
			alg := jwt.GetSigningMethod(claims["alg"].(string))
			// log.I.S(alg)
			token := jwt.NewWithClaims(alg, claims)
			// log.I.S(token)
			var signed string
			if signed, err = token.SignedString(sec); chk.E(err) {
				fail(err.Error())
			}
			fmt.Printf("Authorization: Bearer %s\n", signed)
		}
	}
}
