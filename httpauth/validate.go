package httpauth

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/tag"
)

// ValidateRequest verifies a received http.Request has got a valid
// authentication event or token in it, and provides the public key that should be
// verified to be authorized to access the resource associated with the request.
//
// A VerifyJWTFunc should be provided in order to search the event store for a
// kind 13004 with a JWT signer pubkey that is granted authority for the request.
func ValidateRequest(r *http.Request, vfn VerifyJWTFunc) (valid bool, pubkey []byte, err error) {
	log.I.F("validating nip-98")
	val := r.Header.Get(HeaderKey)
	if val == "" {
		err = errorf.E("'%s' key missing from request header", HeaderKey)
		return
	}
	switch {
	case strings.HasPrefix(val, NIP98Prefix):
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
		if ev.Tags.Len() < 2 {
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
		proto := r.URL.Scheme
		// if this came through a proxy we need to get the protocol to match the event
		if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
			proto = p
		}
		if proto == "" {
			proto = "http"
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
	case strings.HasPrefix(val, JWTPrefix):
		if vfn != nil {
			err = errorf.E("JWT bearer header found but no JWT verifier function provided")
			return
		}
		split := strings.Split(val, " ")
		if len(split) == 1 {
			err = errorf.E("missing JWT auth token from '%s' http header key: '%s'",
				HeaderKey, val)
		}
		if len(split) > 2 {
			err = errorf.E("extraneous content after second field space separated: %s", val)
			return
		}
		// The u tag MUST be exactly the same as the absolute request URL (including query parameters).
		proto := r.URL.Scheme
		// if this came through a proxy we need to get the protocol to match the event
		if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
			proto = p
		}
		if proto == "" {
			proto = "http"
		}
		fullUrl := proto + "://" + r.Host + r.URL.RequestURI()
		if pubkey, valid, err = VerifyJWTtoken(split[1], fullUrl, vfn); chk.E(err) {
			return
		}

	default:
		err = errorf.E("invalid '%s' value: '%s'", HeaderKey, val)
		return
	}

	return
}
