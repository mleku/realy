package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/p256k"
)

const secEnv = "NOSTR_SECRET_KEY"

var userAgent = fmt.Sprintf("curdl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`curdl help:

to read:

    curdl post <url>

output will be rendered to stdout.

to write:

    curdl post <url> [<payload sha256 hash in hex>] <file>

    if payload hash is not given, it is not computed. NIP-98 authentication can optionally require the file upload hash be in the "payload" HTTP header with the value as the hash encoded in hexadecimal.

for nostr http protocol:

	curdl nostr <url> <nostr http json>
`)
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
	case "get", "post":
	default:
		fail("first parameter must be either 'get' or 'post', got '%s'", meth)
	}
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
	// we assume the hash comes before the filename if it is generated using sha256sum
	var req *http.Request
	var payload io.ReadCloser
	contentLength := int64(math.MaxInt64)
	switch meth {
	case "nostr":
		fail("nostr json protocol not yet implemented")
	case "post":
		// get the file path parameters and optional hash
		var filePath, h string
		if len(os.Args) == 4 {
			filePath = os.Args[3]
		} else if len(os.Args) == 5 {
			// only need to check this is hex
			if _, err = hex.Dec(os.Args[3]); chk.E(err) {
				// if it's not hex and there is 4 args then this is invalid
				fail("invalid missing hex in parameters with 4 parameters set: %v", os.Args[1:])
			}
			filePath = os.Args[4]
			h = os.Args[3]
		} else {
			fail("extraneous stuff in commandline: %v", os.Args[3:])
		}
		log.I.F("reading from %s optional hash: %s", filePath, h)
		var fi os.FileInfo
		if fi, err = os.Stat(filePath); chk.E(err) {
			return
		}
		contentLength = fi.Size()
		if payload, err = os.Open(filePath); chk.E(err) {
			return
		}
		log.I.F("opened file %s", filePath)
		var r *http.Request
		if r, err = httpauth.MakePostRequest(ur, h, userAgent, sign, payload, contentLength); chk.E(err) {
			fail(err.Error())
		}
		r.GetBody = func() (rc io.ReadCloser, err error) {
			rc = payload
			return
		}
		log.I.S(r)
		client := &http.Client{}
		var res *http.Response
		if res, err = client.Do(r); chk.E(err) {
			return
		}
		log.I.S(res)
		defer res.Body.Close()
		if io.Copy(os.Stdout, res.Body); chk.E(err) {
			return
		}

	case "get":
		req, err = httpauth.MakeGetRequest(ur, userAgent, sign)
		client := &http.Client{
			CheckRedirect: func(req *http.Request,
				via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		var res *http.Response
		if res, err = client.Do(req); chk.E(err) {
			err = errorf.E("request failed: %w", err)
			return
		}
		defer res.Body.Close()
		if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
			return
		}
	}
}