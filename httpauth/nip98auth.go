package httpauth

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/signer"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

const (
	HeaderKey   = "Authorization"
	NIP98Prefix = "Nostr"
)

func MakeNIP98Event(u, method string) (ev *event.T) {
	ev = &event.T{
		CreatedAt: timestamp.Now(),
		Kind:      kind.HTTPAuth,
		Tags:      tags.New(tag.New("u", u), tag.New("method", strings.ToUpper(method))),
	}
	return
}

// MakeNIP98PostRequest creates a new http request with a nip-98 authentication code
// in the header, signed by a provided signer.I with secret loaded, for pushing
// data up to a server.
func MakeNIP98PostRequest(ur *url.URL, payloadHash, userAgent string,
	sign signer.I, payload io.ReadCloser, contentLength int64) (r *http.Request, err error) {

	const method = "POST"
	ev := MakeNIP98Event(ur.String(), method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
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
	r.Header.Add(HeaderKey, "Nostr "+b64)
	if payloadHash != "" {
		r.Header.Add("payload", payloadHash)
	}
	// log.I.F("Authorization: %s", req.Header.Get("Authorization"))
	r.Header.Add("User-Agent", userAgent)
	// r.Header.Add("Content-Type", "application/binary")
	log.I.F("made post request")
	return
}

// MakeNIP98GetRequest creates a new http request with a nip-98 authentication code
// in the header, signed by a provided signer.I with secret loaded. This is for
// a simple query on a path and parameters.
func MakeNIP98GetRequest(u *url.URL, userAgent string, sign signer.I) (r *http.Request,
	err error) {

	const method = "GET"
	ev := MakeNIP98Event(u.String(), method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	if r, err = http.NewRequest(method, u.String(), nil); chk.E(err) {
		return
	}
	r.Header.Add(HeaderKey, "Nostr "+b64)
	// log.I.F("Authorization: %s", req.Header.Get("Authorization"))
	r.Header.Add("User-Agent", userAgent[:len(userAgent)-1])
	// r.Header.Add("Content-Type", "application/text")
	return
}
