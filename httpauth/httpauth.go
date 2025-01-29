package httpauth

import (
	"realy.lol/event"
	"realy.lol/tags"
	"realy.lol/tag"
	"realy.lol/kind"
	"strings"
	"net/http"
	"realy.lol/signer"
	"net/url"
	"encoding/base64"
	"io"
)

func MakeEvent(u, method st) (ev *event.T) {
	ev = &event.T{
		Kind: kind.HTTPAuth,
		Tags: tags.New(tag.New("u", u), tag.New("method", strings.ToUpper(method))),
	}
	return
}

func MakeRequest(ur, meth st,
	sign signer.I, payload io.ReadCloser) (r *http.Request, err er) {

	if _, err = url.Parse(ur); chk.E(err) {
		return
	}
	method := strings.ToUpper(meth)
	ev := MakeEvent(ur, method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.I.F("%s", ev.Serialize())
	b64 := base64.RawURLEncoding.EncodeToString(ev.Serialize())
	log.I.F("%s", b64)
	var req *http.Request
	if req, err = http.NewRequest(method, ur, nil); chk.E(err) {
		return
	}
	req.Header.Add("Authorization", "Nostr "+b64)
	switch method {
	case "POST":
		// add the reader for the data
		req.Body = payload
	case "GET":
	default:
		err = errorf.E("unsupported http method: %s", method)
		return
	}
	return
}
