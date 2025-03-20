package realy

import (
	"realy.lol/auth"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/web"
)

func (s *Server) handleAuth(ws *web.Socket, req []byte) (msg []byte) {
	if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
		svcUrl := auther.ServiceUrl(ws.Req())
		if svcUrl == "" {
			return
		}
		log.T.F("received auth response,%s", req)
		var err error
		var rem []byte
		env := authenvelope.NewResponse()
		if rem, err = env.Unmarshal(req); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.F("extra '%s'", rem)
		}
		var valid bool
		if valid, err = auth.Validate(env.Event, []byte(ws.Challenge()), svcUrl); chk.E(err) {
			e := err.Error()
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Error.F(e)
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F("failed to authenticate")).Write(ws); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.ID, true, []byte{}).Write(ws); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", ws.RealRemote(), env.Event.PubKey)
			ws.SetAuthed(string(env.Event.PubKey))
		}
	}
	return
}
