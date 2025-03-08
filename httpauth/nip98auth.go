package httpauth

import (
	"encoding/base64"
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

func AddNIP98Header(r *http.Request, ur *url.URL, method string, sign signer.I) (err error) {
	ev := MakeNIP98Event(ur.String(), method)
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	r.Header.Add(HeaderKey, "Nostr "+b64)
	return
}
