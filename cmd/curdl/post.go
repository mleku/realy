package main

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/signer"
)

func Post(args []string, ur *url.URL, sign signer.I) (err error) {
	var contentLength int64
	var payload io.ReadCloser
	// get the file path parameters and optional hash
	var filePath, h string
	if len(args) == 4 {
		filePath = args[3]
	} else if len(args) == 5 {
		// only need to check this is hex
		if _, err = hex.Dec(args[3]); chk.E(err) {
			// if it's not hex and there is 4 args then this is invalid
			fail("invalid missing hex in parameters with 4 parameters set: %v", args[1:])
		}
		filePath = args[4]
		h = args[3]
	} else {
		fail("extraneous stuff in commandline: %v", args[3:])
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
