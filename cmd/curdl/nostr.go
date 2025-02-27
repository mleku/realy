package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

func Nostr(args []string, ur *url.URL) (err error) {
	switch ur.Path {
	case "/relayinfo": // put all get methods here
		var r *http.Request
		if r, err = http.NewRequest("GET", ur.String(), nil); chk.E(err) {
			fail(err.Error())
		}
		r.Header.Add("User-Agent", userAgent)
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
	return
}
