package main

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"realy.lol/httpauth"
	"realy.lol/signer"
)

func Get(ur *url.URL, sign signer.I) (err error) {
	var r *http.Request
	if r, err = httpauth.MakeNIP98GetRequest(ur, userAgent, sign); chk.E(err) {
		fail(err.Error())
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
	defer res.Body.Close()
	if _, err = io.Copy(os.Stdout, res.Body); chk.E(err) {
		return
	}
	return
}
