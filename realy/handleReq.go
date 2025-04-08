package realy

import (
	"bytes"
	"errors"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/context"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closedenvelope"
	"realy.lol/envelopes/eoseenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/reqenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/normalize"
	"realy.lol/realy/pointers"
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

func (s *Server) handleReq(c context.T, ws *web.Socket, req []byte, sto store.I) (r []byte) {
	var err error
	var rem []byte
	env := reqenvelope.New()
	if rem, err = env.Unmarshal(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	allowed := env.Filters
	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		var accepted, modified bool
		allowed, accepted, modified = accepter.AcceptReq(c, ws.Req(), env.Subscription.T,
			env.Filters,
			[]byte(ws.Authed()))
		if !accepted || allowed == nil || modified {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
				}
				log.T.F("requesting auth from client from %s, challenge '%s'",
					ws.RealRemote(), ws.Challenge())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				if !modified {
					return
				}
			}
		}
	}
	// log.I.ToSliceOfBytes("handling %s", env.Marshal(nil))
	if allowed != env.Filters {
		defer func() {
			var auther relay.Authenticator
			var ok bool
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() &&
				!ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
				}
				log.T.F("requesting auth from client from %s, challenge '%s'", ws.RealRemote(),
					ws.Challenge())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}()
	}
	if allowed == nil {
		return
	}
	for _, f := range allowed.F {
		var i uint
		if pointers.Present(f.Limit) {
			if *f.Limit == 0 {
				continue
			}
			i = *f.Limit
		}
		var auther relay.Authenticator
		var ok bool
		if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.IsPrivileged() {
				log.T.F("privileged request\n%s", f.Serialize())
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("#p"))
				switch {
				case len(ws.Authed()) == 0:
					// ws.RequestAuth()
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for processing request due to presence of privileged kinds (DMs, app specific data)")).Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					notice := normalize.Restricted.F("this realy does not serve DMs or Application Specific Data " +
						"to unauthenticated users or to npubs not found in the event tags or author fields, does your " +
						"client implement NIP-42?")
					return notice
				case senders.Contains(ws.AuthedBytes()) ||
					receivers.ContainsAny([]byte("#p"), tag.New(ws.AuthedBytes())):
					log.T.F("user %0x from %s allowed to query for privileged event",
						ws.AuthedBytes(), ws.RealRemote())
				default:
					return normalize.Restricted.F("authenticated user %0x does not have authorization for "+
						"requested filters", ws.AuthedBytes())
				}
			}
		}
		var events event.Ts
		log.D.F("query from %s %0x,%s", ws.RealRemote(), ws.AuthedBytes(), f.Serialize())
		if events, err = sto.QueryEvents(c, f); err != nil {
			log.E.F("eventstore: %v", err)
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
			continue
		}
		aut := ws.AuthedBytes()
		// remove events from muted authors if we have the authed user's mute list.
		if ws.IsAuthed() {
			var mutes event.Ts
			if mutes, err = sto.QueryEvents(c, &filter.T{Authors: tag.New(aut),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs [][]byte
				for _, ev := range mutes {
					for _, t := range ev.Tags.ToSliceOfTags() {
						if bytes.Equal(t.Key(), []byte("p")) {
							var p []byte
							if p, err = hex.Dec(string(t.Value())); chk.E(err) {
								continue
							}
							mutePubs = append(mutePubs, p)
						}
					}
				}
				var tmp event.Ts
				for _, ev := range events {
					for _, pk := range mutePubs {
						if bytes.Equal(ev.Pubkey, pk) {
							continue
						}
						tmp = append(tmp, ev)
					}
				}
				// remove privileged events
				events = tmp
			}
		}
		// remove privileged events as they come through in scrape queries
		var tmp event.Ts
		for _, ev := range events {
			receivers := f.Tags.GetAll(tag.New("#p"))
			// if auth is required, kind is privileged and there is no authed pubkey, skip
			if s.authRequired && ev.Kind.IsPrivileged() && len(aut) == 0 {
				// log.I.ToSliceOfBytes("skipping event because event kind is %d and no auth", ev.Kind.K)
				if auther != nil {
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for processing request due to presence of privileged kinds (DMs, app specific data)")).Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					notice := normalize.Restricted.F("this realy does not serve DMs or Application Specific Data " +
						"to unauthenticated users or to npubs not found in the event tags or author fields, does your " +
						"client implement NIP-42?")
					return notice
				}
				continue
			}
			// if the authed pubkey is not present in the pubkey or p tags, skip
			if ev.Kind.IsPrivileged() && (!bytes.Equal(ev.Pubkey, aut) ||
				!receivers.ContainsAny([]byte("#p"), tag.New(ws.AuthedBytes()))) {
				// log.I.ToSliceOfBytes("skipping event %0x because authed key %0x is in neither pubkey or p tag",
				// 	ev.Id, aut)
				if auther != nil {
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for processing request due to presence of privileged kinds (DMs, app specific data)")).Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					notice := normalize.Restricted.F("this realy does not serve DMs or Application Specific Data " +
						"to unauthenticated users or to npubs not found in the event tags or author fields, does your " +
						"client implement NIP-42?")
					return notice
				}
				continue
			}
			tmp = append(tmp, ev)
		}
		events = tmp
		// write out the events to the socket
		for _, ev := range events {
			if s.options.SkipEventFunc != nil && s.options.SkipEventFunc(ev) {
				continue
			}
			i--
			if i < 0 {
				break
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(env.Subscription.T, ev); chk.E(err) {
				return
			}
			if err = res.Write(ws); chk.E(err) {
				return
			}
		}
	}
	if err = eoseenvelope.NewFrom(env.Subscription).Write(ws); chk.E(err) {
		return
	}
	if env.Filters != allowed {
		return
	}
	s.Listeners.Set(env.Subscription.String(), ws, env.Filters)
	return
}
