// Package main is a simple implementation of a cURL like tool that can do
// simple GET/POST operations on a HTTP server that understands NIP-98
// authentication, with the signing key found in an environment variable.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/p256k"
	"realy.lol/sha256"
	"realy.lol/signer"
)

const secEnv = "NOSTR_SECRET_KEY"

var userAgent = fmt.Sprintf("nurl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	// lol.SetLogLevel("trace")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`nurl help:

for nostr http using NIP-98 HTTP authentication:

    nurl <url> <file>

	if no file is given, the request will be processed as a HTTP GET (if relevant there can be request parameters).

	* NIP-98 secret will be expected in the environment variable "%s" - if absent, will not be added to the header. Endpoint is assumed to not require it if absent. An error will be returned if it was needed.

	output will be rendered to stdout

`, secEnv)
		os.Exit(0)
	}
	if len(os.Args) < 2 {
		fail(`error: nurl requires minimum 1 arg:  <url> 

    signing nsec (in bech32 format) is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	var err error
	var sign signer.I
	if sign, err = GetNIP98Signer(); err != nil {
	}
	var ur *url.URL
	if ur, err = url.Parse(os.Args[1]); chk.E(err) {
		fail("invalid URL: `%s` error: `%s`", os.Args[2], err.Error())
	}
	log.T.S(ur)
	if len(os.Args) == 2 {
		if err = Get(ur, sign); chk.E(err) {
			fail(err.Error())
		}
		return
	}
	if err = Post(os.Args[2], ur, sign); chk.E(err) {
		fail(err.Error())
	}
}

func GetNIP98Signer() (sign signer.I, err error) {
	nsex := os.Getenv(secEnv)
	var sk []byte
	if len(nsex) == 0 {
		err = errorf.E("no bech32 secret key found in environment variable %s", secEnv)
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

func Get(ur *url.URL, sign signer.I) (err error) {
	log.T.F("GET")
	var r *http.Request
	if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
		return
	}
	r.Header.Add("User-Agent", userAgent)
	r.Header.Add("Accept", "application/nostr+json")
	if sign != nil {
		if err = httpauth.AddNIP98Header(r, ur, "GET", "", sign, 0); chk.E(err) {
			fail(err.Error())
		}
	}
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
	if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
		res.Body.Close()
		return
	}
	res.Body.Close()
	return
}

func Post(f string, ur *url.URL, sign signer.I) (err error) {
	log.T.F("POST")
	var contentLength int64
	var payload io.ReadCloser
	// get the file path parameters and optional hash
	var fi os.FileInfo
	if fi, err = os.Stat(f); chk.E(err) {
		return
	}
	var b []byte
	if b, err = os.ReadFile(f); chk.E(err) {
		return
	}
	hb := sha256.Sum256(b)
	h := hex.Enc(hb[:])
	contentLength = fi.Size()
	if payload, err = os.Open(f); chk.E(err) {
		return
	}
	log.T.F("opened file %s hash %s", f, h)
	var r *http.Request
	r = &http.Request{
		Method:        "POST",
		URL:           ur,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          payload,
		ContentLength: contentLength,
		Host:          ur.Host,
	}
	r.Header.Add("User-Agent", userAgent)
	r.Header.Add("Accept", "application/nostr+json")
	if sign != nil {
		if err = httpauth.AddNIP98Header(r, ur, "POST", h, sign, 0); chk.E(err) {
			fail(err.Error())
		}
	}
	r.GetBody = func() (rc io.ReadCloser, err error) {
		rc = payload
		return
	}
	// log.I.S(r)
	client := &http.Client{}
	var res *http.Response
	if res, err = client.Do(r); chk.E(err) {
		return
	}
	// log.I.S(res)
	defer res.Body.Close()
	if io.Copy(os.Stdout, res.Body); chk.E(err) {
		return
	}
	fmt.Println()
	return
}
