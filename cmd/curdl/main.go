package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/lol"
	"realy.lol/p256k"
	"realy.lol/signer"
)

const secEnv = "NOSTR_SECRET_KEY"
const jwtSecEnv = "NOSTR_JWT_SECRET"

var userAgent = fmt.Sprintf("curdl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	lol.SetLogLevel("trace")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`curdl help:

for nostr http using NIP-98 HTTP authentication:

    curdl req <url> [[<payload sha256 hash in hex>] <file>]

	if no file or hash is given, the request will be processed as a HTTP GET (if relevant there can be request parameters).

    if payload hash is not given, it is not computed. NIP-98 authentication can optionally require the file upload hash be in the "payload" HTTP header with the value as the hash encoded in hexadecimal, if the relay requires this, use "sha256sum <file>" in place of the last two parameters for this result, as it may refuse to process it without it.

	NIP-98 secret will be expected in the environment variable "%s" - if absent, will not be added to the header.

	output will be rendered to stdout.

for nostr http methods using JWT:

	curdl jwt <url> [[<payload sha256 hash in hex>] <file>]

    if no file (or file and hash) is given, it will be executed as a GET request (if relevant there can be request parameters).

	JWT secret will be expected in  the environment variable "%s" - if absent, will not be added to the header. This variable consists of two fields, <JWT token in base64>,<authorized pubkey in hex>

	output will be rendered to stdout.
`, secEnv, jwtSecEnv)
		os.Exit(0)
	}
	if len(os.Args) < 3 {
		fail(`error: curdl requires minimum 2 args: <get> <url> 

    signing nsec (in bech32 format) is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	meth := strings.ToLower(os.Args[1])
	var err error
	var ur *url.URL
	if ur, err = url.Parse(os.Args[2]); chk.E(err) {
		fail("invalid URL: `%s` error: `%s`", os.Args[2], err.Error())
	}
	switch meth {
	case "req", "jwt":
	default:
		fail("first parameter must be either 'req', or 'jwt', got '%s'", meth)
	}
	switch meth {
	case "jwt":
		if err = NostrJWT(os.Args, ur, "", os.Getenv(jwtSecEnv)); chk.E(err) {
			fail(err.Error())
		}

	case "req":
		var sign signer.I
		if sign, err = GetNIP98Signer(); chk.E(err) {
			fail(err.Error())
		}
		if len(os.Args) == 3 {
			if err = Get(ur, sign); chk.E(err) {
				fail(err.Error())
			}
			return
		}
		if err = Post(os.Args, ur, sign); chk.E(err) {
			fail(err.Error())
		}
	}
}

func GetNIP98Signer() (sign signer.I, err error) {
	nsex := os.Getenv(secEnv)
	var sk []byte
	if len(nsex) == 0 {
		err = errorf.E("no key found in environment variable %s", secEnv)
		return
	} else if sk, err = bech32encoding.NsecToBytes([]byte(nsex)); chk.E(err) {
		err = errorf.E("failed to decode nsec: '%s'", err.Error())
		return
	}
	sign = &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		err = errorf.E("failed to init signer: '%s'", err.Error())
		return
	}
	return
}
