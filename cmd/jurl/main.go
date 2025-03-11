package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	realy_lol "realy.lol"
	"realy.lol/httpauth"
	"realy.lol/lol"
	"realy.lol/sha256"
)

const issuer = "NOSTR_PUBLIC_KEY"
const secEnv = "NOSTR_JWT_SECRET"

var userAgent = fmt.Sprintf("nurl/%s", realy_lol.Version)

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	lol.SetLogLevel("trace")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`jurl help:

for nostr http using JWT HTTP authentication:

    jurl <url> <file>

	if no file is given, the request will be processed as a HTTP GET 
	(if relevant there can be request parameters)

	* JWT secret will be expected in the environment variable "%s" - 
	if absent, will not be added to the header

	* the "issuer" key, the nostr public key associated with the JWT public 
	key must also be available at %s or authorization will fail

	Endpoint is assumed to not require it if absent. an error will be returned 
	if it was needed; if the relay does not have a kind 13004 event binding the 
	JWT public key to a nostr public key then it will also fail

	output will be rendered to stdout

* 	note this tool is designed to generate momentary authorizations, if you have 
	an unexpired JWT token you can just add it with the 

		"Authorization: Bearer <token>" 

	as an additional HTTP header field if you have it to any other HTTP request 
	tool, such as "curl", by using "nostrjwt bearer" command

`, secEnv, issuer)
		os.Exit(0)
	}
	if len(os.Args) < 2 {
		fail(`error: nurl requires minimum 1 arg:  <url> 

    signing nsec (in bech32 format) is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	if len(os.Args) < 2 {
		fail(`error: nurl requires minimum 1 arg:  <url> 

    signing JWT secret is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	var err error
	var ur *url.URL
	if ur, err = url.Parse(os.Args[1]); chk.E(err) {
		fail("invalid URL: `%s` error: `%s`", os.Args[2], err.Error())
	}
	jwtSec := os.Getenv(secEnv)
	bearer := os.Getenv(issuer)
	if jwtSec == "" {
		log.I.F("no key found in environment variable %s", secEnv)
	} else {
		var jskb []byte
		if jskb, err = base64.URLEncoding.DecodeString(jwtSec); chk.E(err) {
			fail(err.Error())
		}
		var sec *ecdsa.PrivateKey
		if sec, err = x509.ParseECPrivateKey(jskb); chk.E(err) {
			fail(err.Error())
		}
		var claims []byte
		// generate claim
		if claims, err = httpauth.GenerateJWTClaims(bearer, ur.String()); chk.E(err) {
			fail(err.Error())
		}
		if bearer, err = httpauth.SignJWTtoken(claims, sec); chk.E(err) {
			fail(err.Error())
		}
	}
	if len(os.Args) == 2 {
		if err = Get(ur, bearer); chk.E(err) {
			fail(err.Error())
		}
		return
	}
	if err = Post(os.Args[2], ur, bearer); chk.E(err) {
		fail(err.Error())
	}
}

func Get(ur *url.URL, bearer string) (err error) {
	var r *http.Request
	if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
		return
	}
	r.Header.Add("User-Agent", userAgent)
	if bearer != "" {
		r.Header.Add("Authorization", "Authorization "+bearer)
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

func Post(filePath string, ur *url.URL, bearer string) (err error) {
	var contentLength int64
	var payload io.ReadCloser
	var b []byte
	if b, err = os.ReadFile(filePath); chk.E(err) {
		fail(err.Error())
	}
	H := sha256.Sum256(b)
	_ = H
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
	r.Header.Add("Authorization", "Bearer "+bearer)
	r.Header.Add("Accept", "application/nostr+json")
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
