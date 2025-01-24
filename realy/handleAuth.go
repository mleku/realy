package realy

import (
	"realy.lol/auth"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/web"
)

func (s *Server) handleAuth(ws *web.Socket, req by) (msg by) {
	if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
		svcUrl := auther.ServiceUrl(ws.Req())
		if svcUrl == "" {
			return
		}
		log.T.F("received auth response,%s", req)
		var err er
		var rem by
		env := authenvelope.NewResponse()
		if rem, err = env.Unmarshal(req); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.F("extra '%s'", rem)
		}
		var valid bo
		if valid, err = auth.Validate(env.Event, by(ws.Challenge()), svcUrl); chk.E(err) {
			e := err.Error()
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
				return by(err.Error())
			}
			return normalize.Error.F(e)
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F("failed to authenticate")).Write(ws); chk.E(err) {
				return by(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.ID, true, by{}).Write(ws); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", ws.RealRemote(), env.Event.PubKey)
			ws.SetAuthed(st(env.Event.PubKey))
			if s.relay.NoLimiter(env.Event.PubKey) {
				// if user is authed as a direct follow of the owners' follow list this means
				// they are paying or guest users on the relay and there is no limit on their
				// request/publish rates.
				//
				// todo: maybe this should be more stringent and relax the limiter but for now
				//  there is no per user access accounting or any such thing. there may be need
				//  of it. the reason being that publishing large numbers of document events is
				//  a projected use case.
				ws.SetLimiter(nil)
			}
		}
	}
	return
}
