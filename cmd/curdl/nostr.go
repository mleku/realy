package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

func NostrJWT(args []string, ur *url.URL, jwtSec, pubkey string) (err error) {
	var r *http.Request
	var res *http.Response
	var client *http.Client
	if len(args) == 3 {
		// this is a GET request
		if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
			fail(err.Error())
		}
		r.Header.Add("User-Agent", userAgent)
		r.Header.Add("Accept", "application/nostr+json")
		client = &http.Client{
			CheckRedirect: func(req *http.Request,
				via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		if res, err = client.Do(r); chk.E(err) {
			err = errorf.E("request failed: %w", err)
			return
		}
		defer res.Body.Close()
		if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
			return
		}
		return
	}
	// this is a POST request
	if r, err = http.NewRequest("POST", ur.String(), os.Stdin); chk.E(err) {
		fail(err.Error())
	}
	r.Header.Add("User-Agent", userAgent)
	r.Header.Add("Accept", "application/nostr+json")
	client = &http.Client{
		CheckRedirect: func(req *http.Request,
			via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if res, err = client.Do(r); chk.E(err) {
		err = errorf.E("request failed: %w", err)
		return
	}
	defer res.Body.Close()
	if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
		return
	}

	return
}
