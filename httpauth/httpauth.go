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

func MakeRequest(ur, meth string,
	sign signer.I, payloadHash string, payload ...io.ReadCloser) (r *http.Request, err error) {

	if _, err = url.Parse(ur); chk.E(err) {
		return
	}
	method := strings.ToUpper(meth)
	ev := MakeEvent(ur, method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	// log.I.F("signing event %s", ev.Serialize())
	log.T.F("nip-98 auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	if r, err = http.NewRequest(method, ur, nil); chk.E(err) {
		return
	}

	r.Header.Add(HeaderKey, "Nostr "+b64)
	if payloadHash != "" {
		r.Header.Add("payload", payloadHash)
	}
	// log.I.F("Authorization: %s", req.Header.Get("Authorization"))
	switch method {
	case "POST":
		// add the reader for the data
		if len(payload) < 1 {
			r.Body = payload[0]
		}
	case "GET":

	default:
		err = errorf.E("unsupported http method: %s", method)
		return
	}
	return
}

func ValidateRequest(r *http.Request) (valid bool, pubkey []byte, err error) {
	val := r.Header.Get(HeaderKey)
	if val == "" {
		err = errorf.E("'%s' key missing from request header")
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
	if r.URL.String() != string(uts[0].Value()) {
		err = errorf.E("request has URL %s but signed nip-98 event has url %s",
			r.URL.String(), string(uts[0].Value()))
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
