package httpauth

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/log"
	"realy.lol/signer"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

const (
	HeaderKey   = "Authorization"
	NIP98Prefix = "Nostr"
)

// MakeNIP98Event creates a new NIP-98 event. If expiry is given, method is ignored, otherwise
// either option is the same.
func MakeNIP98Event(u, method, hash string, expiry int64) (ev *event.T) {
	var t []*tag.T
	t = append(t, tag.New("u", u))
	if expiry > 0 {
		t = append(t,
			tag.New("expiration", timestamp.FromUnix(expiry).String()))
	} else {
		t = append(t,
			tag.New("method", strings.ToUpper(method)))
	}
	if hash != "" {
		t = append(t, tag.New("payload", hash))
	}
	ev = &event.T{
		CreatedAt: timestamp.Now(),
		Kind:      kind.HTTPAuth,
		Tags:      tags.New(t...),
	}
	return
}

// AddNIP98Header creates a NIP-98 http auth event and adds the standard header to a provided
// http.Request.
func AddNIP98Header(r *http.Request, ur *url.URL, method, hash string,
	sign signer.I, expiry int64) (err error) {
	ev := MakeNIP98Event(ur.String(), method, hash, expiry)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	r.Header.Add(HeaderKey, "Nostr "+b64)
	return
}
