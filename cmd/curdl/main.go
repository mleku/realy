package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/lol"
	"realy.lol/p256k"
)

const secEnv = "NOSTR_SECRET_KEY"

var userAgent = fmt.Sprintf("curdl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	lol.SetLogLevel("trace")
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
	case "get", "post", "nostr":
	default:
		fail("first parameter must be either 'get', 'post', or 'nostr', got '%s'", meth)
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
	switch meth {
	case "nostr":
		switch ur.Path {
		case "/relayinfo":
			var r *http.Request
			if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
				fail(err.Error())
			}
			r.Header.Add("User-Agent", userAgent[:len(userAgent)-1])
			r.Header.Add("Accept", "application/nostr+json")
			client := &http.Client{
				CheckRedirect: func(req *http.Request,
					via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			var res *http.Response
			if res, err = client.Do(r); chk.E(err) {
				err = errorf.E("request failed: %w", err)
				return
			}
			defer res.Body.Close()
			if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
				return
			}
		default:
			fail("unrecognised method '%s'")
		}

	case "post":
		if err = Post(os.Args, ur, sign); chk.E(err) {
			fail(err.Error())
		}

	case "get":
		if err = Get(ur, sign); chk.E(err) {
			fail(err.Error())
		}
	}
}
