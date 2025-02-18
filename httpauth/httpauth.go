package httpauth

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/signer"
	"realy.lol/tag"
	"realy.lol/tags"
)

const (
	HeaderKey    = "Authorization"
	HeaderPrefix = "Nostr"
)

func MakeEvent(u, method string) (ev *event.T) {
	ev = &event.T{
		Kind: kind.HTTPAuth,
		Tags: tags.New(tag.New("u", u), tag.New("method", strings.ToUpper(method))),
	}
	return
}

// MakePostRequest creates a new http request with a nip-98 authentication code
// in the header, signed by a provided signer.I with secret loaded, for pushing
// data up to a server.
func MakePostRequest(ur *url.URL, payloadHash, userAgent string,
	sign signer.I, payload io.ReadCloser) (r *http.Request, err error) {

	const method = "POST"
	ev := MakeEvent(ur.String(), method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	// log.I.F("signing event %s", ev.Serialize())
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	r = &http.Request{
		Method:     "POST",
		URL:        ur,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       payload,
		Host:       ur.Host,
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

// MakeGetRequest creates a new http request with a nip-98 authentication code
// in the header, signed by a provided signer.I with secret loaded. This is for
// a simple query on a path and parameters.
func MakeGetRequest(u *url.URL, userAgent string, sign signer.I) (r *http.Request,
	err error) {

	const method = "GET"
	log.I.S(u.String())
	ev := MakeEvent(u.String(), method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	if r, err = http.NewRequest(method, u.String(), nil); chk.E(err) {
		return
	}
	log.I.S(r.URL.String())
	r.Header.Add(HeaderKey, "Nostr "+b64)
	// log.I.F("Authorization: %s", req.Header.Get("Authorization"))
	r.Header.Add("User-Agent", userAgent[:len(userAgent)-1])
	// r.Header.Add("Content-Type", "application/text")
	return
}

// ValidateRequest verifies a received http.Request has got a valid
// authentication event in it, and provides the public key that should be
// verified to be authorized to access the resource associated with the request.
func ValidateRequest(r *http.Request) (valid bool, pubkey []byte, err error) {
	val := r.Header.Get(HeaderKey)
	if val == "" {
		err = errorf.E("'%s' key missing from request header", HeaderKey)
		return
	}
	if !strings.HasPrefix(val, HeaderPrefix) {
		err = errorf.E("invalid '%s' value: '%s'", HeaderKey, val)
		return
	}
	split := strings.Split(val, " ")
	if len(split) == 1 {
		err = errorf.E("missing nip-98 auth event from '%s' http header key: '%s'",
			HeaderKey, val)
	}
	if len(split) > 2 {
		err = errorf.E("extraneous content after second field space separated: %s", val)
		return
	}
	var evb []byte
	if evb, err = base64.URLEncoding.DecodeString(split[1]); chk.E(err) {
		return
	}
	ev := event.New()
	var rem []byte
	if rem, err = ev.Unmarshal(evb); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		err = errorf.E("rem", rem)
		return
	}
	log.T.F("received http auth event:\n%s\n", ev.SerializeIndented())
	// The kind MUST be 27235.
	if !ev.Kind.Equal(kind.HTTPAuth) {
		err = errorf.E("invalid kind %d %s in nip-98 http auth event, require %d %s",
			ev.Kind.K, ev.Kind.Name(), kind.HTTPAuth.K, kind.HTTPAuth.Name())
		return
	}
	// The created_at timestamp MUST be within a reasonable time window (suggestion ~60~ 15 seconds)
	ts := ev.CreatedAt.I64()
	tn := time.Now().Unix()
	if ts < tn-15 || ts > tn+15 {
		err = errorf.E("timestamp %d is more than 15 seconds divergent from now %d",
			ts, tn)
		return
	}
	// we are going to say anything not specified in nip-98 is invalid also, such as extra tags
	if ev.Tags.Len() != 2 {
		err = errorf.E("other than exactly 2 tags found in event\n%s",
			ev.Tags.MarshalTo(nil))
		return
	}
	ut := ev.Tags.GetAll(tag.New("u"))
	if ut.Len() != 1 {
		err = errorf.E("more than one \"u\" tag found: '%s'", ut.MarshalTo(nil))
		return
	}
	uts := ut.Value()
	// The u tag MUST be exactly the same as the absolute request URL (including query parameters).

	// log.I.S(r.Proto, r.Host, r.URL)
	proto := r.URL.Scheme
	// if this came through a proxy we need to get the protocol to match the event
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		proto = p
	}
	fullUrl := proto + "://" + r.Host + r.URL.RequestURI()
	evUrl := string(uts[0].Value())
	// log.I.S(r)
	log.T.F("full URL: %s event u tag value: %s", fullUrl, evUrl)
	if fullUrl != evUrl {
		err = errorf.E("request has URL %s but signed nip-98 event has url %s",
			fullUrl, string(uts[0].Value()))
		return
	}
	// The method tag MUST be the same HTTP method used for the requested resource.
	mt := ev.Tags.GetAll(tag.New("method"))
	if mt.Len() != 1 {
		err = errorf.E("more than one \"method\" tag found: '%s'", mt.MarshalTo(nil))
		return
	}
	mts := mt.Value()
	if strings.ToLower(string(mts[0].Value())) != strings.ToLower(r.Method) {
		err = errorf.E("request has method %s but event has method %s",
			string(mts[0].Value()), r.Method)
		return
	}
	if valid, err = ev.Verify(); chk.E(err) {
		return
	}
	if !valid {
		return
	}
	pubkey = ev.PubKey
	return
}