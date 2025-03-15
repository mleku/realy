package realy

import (
	"bytes"
	"io"
	"net/http"
	"sort"

	"realy.lol/context"
	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/httpauth"
	"realy.lol/relay"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/text"
)

// the handleEvents HTTP endpoint accepts an array containing a list of hex
// encoded event IDs in JSON form, eg
//
//	["<hex>",...]
//
// and either returns a line structured JSON containing one event per line of
// the results, or an OK,false,"reason:..." message
//
// the relay should not inform the client if it has excluded events due to lack
// of required authentication for privileged events, the /api endpoint will have
// a list of event kinds that require auth and if the events contain this and
// auth was not made and if made, does not match a pubkey in the relevant events
// it is simply not returned.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	log.I.F("events")
	var fetcher store.FetchByIds
	var ok bool
	var err error
	if fetcher, ok = s.relay.(store.FetchByIds); ok {
		var pubkey []byte
		if _, pubkey, err = httpauth.CheckAuth(r, s.JWTVerifyFunc); chk.E(err) {
			return
		}
		// if auth is enabled, and either required or not set to public readable, and
		// the client did not auth with an Authorization header, send a HTTP
		// Unauthorized status response.
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok &&
			auther.AuthEnabled() &&
			(s.authRequired || !s.publicReadable) &&
			len(pubkey) < 1 {

			http.Error(w,
				"Authentication required for method", http.StatusUnauthorized)
			return
		}
		var req []byte
		if req, err = io.ReadAll(r.Body); chk.E(err) {
			return
		}
		// unmarshal the request
		var t [][]byte
		if t, req, err = text.UnmarshalHexArray(req, sha256.Size); chk.E(err) {
			return
		}
		if len(req) > 0 {
			log.I.S("extra bytes after hex array:\n%s", req)
		}
		var evs event.Ts
		if evs, err = fetcher.FetchByIds(context.Bg(), t); chk.E(err) {
			return
		}
		// filter out privileged kinds if there is no authed pubkey, auth is enabled but
		// the relay is public readable.
		if auther, ok = s.relay.(relay.Authenticator); ok &&
			auther.AuthEnabled() && s.publicReadable {
			var evTmp event.Ts
			if len(pubkey) < 1 {
				// if not authed, remove all privileged event kinds (user must auth if they want
				// their own DMs
				for _, ev := range evs {
					if !ev.Kind.IsPrivileged() {
						evTmp = append(evTmp, ev)
					}
				}
				evs = evTmp
			} else if len(pubkey) == schnorr.PubKeyBytesLen {
				// if authed, filter out any privileged kinds that don't also contain the authed
				// pubkey in either author or p tags
				for _, ev := range evs {
					if !ev.Kind.IsPrivileged() {
						evTmp = append(evTmp, ev)
					} else {
						var containsPubkey bool
						if ev.Tags != nil {
							containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(pubkey))
						}
						if !bytes.Equal(ev.PubKey, pubkey) || containsPubkey {
							log.I.F("authed user %0x not privileged to receive event\n%s",
								pubkey, ev.Serialize())
						} else {
							// authed pubkey matches either author pubkey or is tagged in privileged event
							evTmp = append(evTmp, ev)
						}
					}
				}
				evs = evTmp
			}
		}
		// sort in descending order (reverse chronological order)
		sort.Slice(evs, func(i, j int) bool {
			return evs[i].CreatedAt.Int() > evs[j].CreatedAt.Int()
		})
		for _, ev := range evs {
			if _, err = w.Write(ev.Marshal(nil)); chk.E(err) {
				return
			}
			// results are jsonl format, one line per event
			if _, err = w.Write([]byte{'\n'}); chk.E(err) {
				return
			}
		}
		http.Error(w, "", http.StatusOK)
	} else {
		http.Error(w, "Method not implemented", NI)
	}
	return
}
