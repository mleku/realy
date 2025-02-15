package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"realy.lol/bech32encoding"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/p256k"
	"realy.lol/sha256"
)

const secEnv = "NOSTR_SECRET_KEY"

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`nurl help:

to read:

    nurl <post> <url>

output will be rendered to stdout.

to write:

    nurl <post> <url> [<payload sha256 hash in hex>] <file>

    use '-' as file to indicate to read from the stdin

    if payload hash is not given, it is not computed. NIP-98 authentication can optionally require the file upload hash be in the "payload" HTTP header with the value as the hash encoded in hexadecimal.
`)
		os.Exit(0)
	}
	if len(os.Args) < 3 {
		fail(`error: nurl requires minimum 2 args: <get/post> <url>

    singing nsec (in bech32 format) is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	meth := strings.ToLower(os.Args[1])
	ur := os.Args[2]
	_ = ur
	switch meth {
	case "get", "post":
	default:
		fail("first parameter must be either 'get' or 'post', got '%s'", meth)
	}
	var sk []byte
	var err error
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
	// log.I.S(sign.Pub())
	var read io.ReadCloser
	// we assume the hash comes before the filename if it is generated using sha256sum
	var h string
	var filename string
	if meth == "post" {
		if len(os.Args) == 4 && len(os.Args[3]) == sha256.Size*2 {
		} else if len(os.Args) == 5 {
			// only need to check this is hex
			if _, err = hex.Dec(os.Args[3]); chk.E(err) {
				// if it's not hex and there is 4 args then this is invalid
				fail("invalid missing hex in parameters with 4 parameters set: %v", os.Args[1:])
			}
			filename = os.Args[4]
			h = os.Args[3]
		} else {
			fail("extraneous stuff in commandline: %v", os.Args[3:])
		}
	}
	// if we are uploading data
	if len(os.Args) > 3 && meth == "post" {
		switch os.Args[3] {
		// as is common, `-` means "read data from stdin"
		case "-":
			read = os.Stdin
		default:
			// otherwise assume it is a file and fail if it isn't
			if read, err = os.OpenFile(filename, os.O_RDONLY, 0600); chk.E(err) {
				fail("failed to open file for reading")
			}
		}
	}
	var r *http.Request
	if r, err = httpauth.MakeRequest(ur, meth, sign, h, read); chk.E(err) {
		fail("failed to create nostr authed http request: %s", err.Error())
	}
	_ = r
}
