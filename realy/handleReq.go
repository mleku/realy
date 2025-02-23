package realy

import (
	"bytes"
	"errors"
	"sort"

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
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

func (s *Server) handleReq(c context.T, ws *web.Socket, req []byte, sto store.I) (r []byte) {
	if !s.publicReadable && (ws.AuthRequested() && len(ws.Authed()) == 0) {
		return []byte("awaiting auth for req")
	}
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
		var accepted bool
		allowed, accepted = accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters,
			[]byte(ws.Authed()))
		if !accepted || allowed == nil {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
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
			return
		}
	}
	// log.I.F("handling %s", env.Marshal(nil))
	if allowed != env.Filters {
		defer func() {
			var auther relay.Authenticator
			var ok bool
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
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
	for _, f := range allowed.F {
		var i uint
		if filter.Present(f.Limit) {
			if *f.Limit == 0 {
				continue
			}
			i = *f.Limit
		}
		// if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
		if f.Kinds.IsPrivileged() {
			log.T.F("privileged request\n%s", f.Serialize())
			senders := f.Authors
			receivers := f.Tags.GetAll(tag.New("#p"))
			switch {
			case len(ws.Authed()) == 0:
				ws.RequestAuth()
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
			case senders.Contains(ws.AuthedBytes()) || receivers.ContainsAny([]byte("#p"),
				tag.New(ws.AuthedBytes())):
				log.T.F("user %0x from %s allowed to query for privileged event",
					ws.AuthedBytes(), ws.RealRemote())
			default:
				return normalize.Restricted.F("authenticated user %0x does not have authorization for "+
					"requested filters", ws.AuthedBytes())
			}
		}
		// }
		var events event.Ts
		log.D.F("query from %s %0x,%s", ws.RealRemote(), ws.AuthedBytes(), f.Serialize())
		if events, err = sto.QueryEvents(c, f); err != nil {
			log.E.F("eventstore: %v", err)
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
			continue
		}
		if aut := ws.Authed(); ws.IsAuthed() {
			var mutes event.Ts
			if mutes, err = sto.QueryEvents(c, &filter.T{Authors: tag.New(aut),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs [][]byte
				for _, ev := range mutes {
					for _, t := range ev.Tags.F() {
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
						if bytes.Equal(ev.PubKey, pk) {
							continue
						}
						tmp = append(tmp, ev)
					}
				}
				events = tmp
			}
		}
		sort.Slice(events, func(i, j int) bool {
			return events[i].CreatedAt.Int() > events[j].CreatedAt.Int()
		})
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
	s.listeners.SetListener(env.Subscription.String(), ws, env.Filters)
	return
}
