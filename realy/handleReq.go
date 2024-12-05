package realy

import (
	"errors"
	"sort"

	"github.com/dgraph-io/badger/v4"

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

func (s *Server) handleReq(c cx, ws *web.Socket, req by, sto store.I) (r by) {
	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		return
	}
	var err er
	var rem by
	env := reqenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	allowed := env.Filters
	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		var accepted bo
		allowed, accepted = accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters,
			by(ws.Authed()))
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
	if allowed != env.Filters {
		defer func() {
			var auther relay.Authenticator
			var ok bo
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
		if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.IsPrivileged() {
				log.T.F("privileged request with auth enabled\n%s", f.Serialize())
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("#p"))
				switch {
				case len(ws.Authed()) == 0:
					ws.RequestAuth()
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					notice := normalize.Restricted.F("this realy does not serve kind-4 to unauthenticated users," + " does your client implement NIP-42?")
					return notice
				case senders.Contains(ws.AuthedBytes()) || receivers.ContainsAny(by("#p"),
					tag.New(ws.AuthedBytes())):
					log.T.F("user %0x from %s allowed to query for privileged event",
						ws.AuthedBytes(), ws.RealRemote())
				default:
					return normalize.Restricted.F("authenticated user %0x does not have"+" authorization for requested filters",
						ws.AuthedBytes())
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
		if aut := ws.Authed(); ws.IsAuthed() {
			var mutes event.Ts
			if mutes, err = sto.QueryEvents(c, &filter.T{Authors: tag.New(aut),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs []by
				for _, ev := range mutes {
					for _, t := range ev.Tags.F() {
						if equals(t.Key(), by("p")) {
							var p by
							if p, err = hex.Dec(st(t.Value())); chk.E(err) {
								continue
							}
							mutePubs = append(mutePubs, p)
						}
					}
				}
				var tmp event.Ts
				for _, ev := range events {
					for _, pk := range mutePubs {
						if equals(ev.PubKey, pk) {
							continue
						}
						tmp = append(tmp, ev)
					}
				}
				events = tmp
			}
		}
		sort.Slice(events, func(i, j int) bo {
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
