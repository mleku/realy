package realy

import (
	"realy.lol/context"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closedenvelope"
	"realy.lol/envelopes/countenvelope"
	"realy.lol/kind"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

func (s *Server) handleCount(c context.T, ws *web.Socket, req by, store store.I) (msg by) {
	counter, ok := store.(relay.EventCounter)
	if !ok {
		return normalize.Restricted.F("this relay does not support NIP-45")
	}
	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		return
	}
	var err er
	var rem by
	env := countenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.Subscription == nil || env.Subscription.String() == "" {
		return normalize.Error.F("COUNT has no <subscription id>")
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
					normalize.AuthRequired.F("auth required for count processing")).Write(ws); chk.E(err) {
				}
				log.I.F("requesting auth from client from %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
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
	var total no
	var approx bo
	if allowed != nil {
		for _, f := range allowed.F {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
				if f.Kinds.Contains(kind.EncryptedDirectMessage) || f.Kinds.Contains(kind.GiftWrap) {
					senders := f.Authors
					receivers := f.Tags.GetAll(tag.New("p"))
					switch {
					case len(ws.Authed()) == 0:
						return normalize.Restricted.F("this realy does not serve kind-4 to unauthenticated users," + " does your client implement NIP-42?")
					case senders.Len() == 1 && receivers.Len() < 2 && equals(senders.F()[0],
						by(ws.Authed())):
					case receivers.Len() == 1 && senders.Len() < 2 && equals(receivers.N(0).Value(),
						by(ws.Authed())):
					default:
						return normalize.Restricted.F("authenticated user does not have" + " authorization for requested filters")
					}
				}
			}
			var count no
			count, approx, err = counter.CountEvents(c, f)
			if err != nil {
				log.E.F("store: %v", err)
				continue
			}
			total += count
		}
	}
	var res *countenvelope.Response
	if res, err = countenvelope.NewResponseFrom(env.Subscription.T, total, approx); chk.E(err) {
		return
	}
	if err = res.Write(ws); chk.E(err) {
		return
	}
	return
}
