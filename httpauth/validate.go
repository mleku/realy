package httpauth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/event"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/tag"

	"realy.lol/log"
)

var ErrMissingKey = fmt.Errorf(
	"'%s' key missing from request header", HeaderKey)

// CheckAuth verifies a received http.Request has got a valid authentication event in it, withan
// optional specification for tolerance of before and after, and provides the public key that
// should be verified to be authorized to access the resource associated with the request.
func CheckAuth(r *http.Request, tolerance ...time.Duration) (valid bool,
	pubkey []byte, err error) {
	val := r.Header.Get(HeaderKey)
	log.I.F(val)
	if val == "" {
		err = ErrMissingKey
		valid = true
		return
	}
	if len(tolerance) == 0 {
		tolerance = append(tolerance, time.Minute)
	}
	log.I.S(tolerance)
	if tolerance[0] == 0 {
		tolerance[0] = time.Minute
	}
	tolerate := int64(tolerance[0] / time.Second)
	log.I.F("validating auth '%s'", val)
	switch {
	case strings.HasPrefix(val, NIP98Prefix):
		log.T.F(val)
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
		// if there is an expiration timestamp it supersedes the created_at for validity.
		exp := ev.Tags.GetAll(tag.New("expiration"))
		if exp.Len() > 1 {
			err = errorf.E("more than one \"expiration\" tag found: '%s'", exp.MarshalTo(nil))
			return
		}
		var expiring bool
		if exp.Len() == 1 {
			ex := ints.New(0)
			exp1 := exp.ToSliceOfTags()[0]
			if rem, err = ex.Unmarshal(exp1.Value()); chk.E(err) {
				return
			}
			tn := time.Now().Unix()
			if tn > ex.Int64()+tolerate {
				err = errorf.E("HTTP auth event is expired %d time now %d",
					tn, ex.Int64()+tolerate)
				return
			}
			expiring = true
		} else {
			// The created_at timestamp MUST be within a reasonable time window (suggestion 60
			// seconds)
			ts := ev.CreatedAt.I64()
			tn := time.Now().Unix()
			if ts < tn-tolerate || ts > tn+tolerate {
				err = errorf.E("timestamp %d is more than %d seconds divergent from now %d",
					ts, tolerate, tn)
				return
			}
		}
		ut := ev.Tags.GetAll(tag.New("u"))
		if ut.Len() > 1 {
			err = errorf.E("more than one \"u\" tag found: '%s'", ut.MarshalTo(nil))
			return
		}
		uts := ut.ToSliceOfTags()
		// The u tag MUST be exactly the same as the absolute request URL (including query
		// parameters).
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
		if expiring {
			// if it is expiring, the URL only needs to be the same prefix to allow its use with
			// multiple endpoints.
			if !strings.HasPrefix(fullUrl, evUrl) {
				err = errorf.E("request URL %s is not prefixed with the u tag URL %s",
					fullUrl, evUrl)
				return
			}
		} else if fullUrl != evUrl {
			err = errorf.E("request has URL %s but signed nip-98 event has url %s",
				fullUrl, string(uts[0].Value()))
			return
		}
		if !expiring {
			// The method tag MUST be the same HTTP method used for the requested resource.
			mt := ev.Tags.GetAll(tag.New("method"))
			if mt.Len() != 1 {
				err = errorf.E("more than one \"method\" tag found: '%s'", mt.MarshalTo(nil))
				return
			}
			mts := mt.ToSliceOfTags()
			if strings.ToLower(string(mts[0].Value())) != strings.ToLower(r.Method) {
				err = errorf.E("request has method %s but event has method %s",
					string(mts[0].Value()), r.Method)
				return
			}
		}
		log.T.F("%d %s", time.Now().Unix(), ev.Serialize())
		if valid, err = ev.Verify(); chk.E(err) {
			return
		}
		if valid {
			log.I.F("event verified %0x", ev.Pubkey)
		}
		if !valid {
			log.T.F("event not verified")
			return
		}
		pubkey = ev.Pubkey
	default:
		err = errorf.E("invalid '%s' value: '%s'", HeaderKey, val)
		return
	}

	return
}
