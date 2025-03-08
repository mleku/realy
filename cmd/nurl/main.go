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
	"realy.lol/lol"
	"realy.lol/p256k"
	"realy.lol/signer"
)

const secEnv = "NOSTR_SECRET_KEY"

var userAgent = fmt.Sprintf("nurl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	lol.SetLogLevel("trace")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`nurl help:

for nostr http using NIP-98 HTTP authentication:

    nurl <url> [[<payload sha256 hash in hex>] <file>]

	if no file or hash is given, the request will be processed as a HTTP GET (if relevant there can be request parameters).

    if payload hash is not given, it is not computed. NIP-98 authentication can optionally require the file upload hash be in the "payload" HTTP header with the value as the hash encoded in hexadecimal, if the relay requires this, use "sha256sum <file>" in place of the last two parameters for this result, as it may refuse to process it without it.

	* NIP-98 secret will be expected in the environment variable "%s" - if absent, will not be added to the header. Endpoint is assumed to not require it. An error will be returned if it was needed.

	output will be rendered to stdout.

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
		// log.I.F()
		// fail(err.Error())
	}
	var ur *url.URL
	if ur, err = url.Parse(os.Args[1]); chk.E(err) {
		fail("invalid URL: `%s` error: `%s`", os.Args[2], err.Error())
	}
	if len(os.Args) == 2 {
		if err = Get(ur, sign); chk.E(err) {
			fail(err.Error())
		}
		return
	}
	if err = Post(os.Args, ur, sign); chk.E(err) {
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
	var r *http.Request
	if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
		return
	}
	r.Header.Add("User-Agent", userAgent)
	if sign != nil {
		if err = httpauth.AddNIP98Header(r, ur, "GET", sign); chk.E(err) {
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

func Post(args []string, ur *url.URL, sign signer.I) (err error) {
	log.I.F("POST")
	var contentLength int64
	var payload io.ReadCloser
	// get the file path parameters and optional hash
	var filePath, h string
	if len(args) == 3 {
		filePath = args[2]
	} else if len(args) == 4 {
		// only need to check this is hex
		if _, err = hex.Dec(args[3]); chk.E(err) {
			// if it's not hex and there is 4 args then this is invalid
			fail("invalid missing hex in parameters with 4 parameters set: %v", args[1:])
		}
		filePath = args[3]
		h = args[2]
	} else {
		fail("extraneous stuff in commandline: %v", args)
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
	if sign != nil {
		if err = httpauth.AddNIP98Header(r, ur, "POST", sign); chk.E(err) {
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

	return
}
