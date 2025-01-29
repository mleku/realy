package main

import (
	"os"
	"fmt"
	"realy.lol/bech32encoding"
	"realy.lol/p256k"
	"realy.lol/httpauth"
	"net/http"
	"strings"
	"io"
)

const secEnv = "NOSTR_SECRET_KEY"

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		fail(`error: nurl requires minimum 2 args: <get/post> <url>

singing nsec (in bech32 format) is expected to be found in %s environment variable.
`, secEnv)
	}
	meth := strings.ToLower(os.Args[1])
	ur := os.Args[2]
	switch meth {
	case "get", "post":
	default:
		fail("first parameter must be either 'get' or 'post', got '%s'", meth)
	}
	var sk []byte
	var err error
	if sk, err = bech32encoding.NsecToBytes([]byte(os.Getenv(secEnv))); chk.E(err) {
		fail("failed to decode nsec: '%s'", err.Error())

	}
	sign := &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		fail("failed to init signer: '%s'", err.Error())
	}
	var read io.ReadCloser
	// if we are uploading data
	if len(os.Args) > 3 && meth == "post" {
		switch os.Args[3] {
		// as is common, `-` means "read data from stdin"
		case "-":
			read = os.Stdin
		default:
			// otherwise assume it is a file and fail if it isn't
			if read, err = os.OpenFile(os.Args[3], os.O_RDONLY, 0600); chk.E(err) {
				fail("failed to open file for reading")
			}
		}
	}
	var r *http.Request
	if r, err = httpauth.MakeRequest(ur, meth, sign, read); chk.E(err) {
		fail("failed to create nostr authed http request: %s", err.Error())
	}
	_ = r
}
