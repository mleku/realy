package realy

import (
	"realy.mleku.dev/auth"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/normalize"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/ws"
)

func (s *Server) handleAuth(ws *ws.Listener, req []byte) (msg []byte) {
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
			if err = okenvelope.NewFrom(env.Event.Id, false,
				normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Error.F(e)
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.Id, false,
				normalize.Error.F("failed to authenticate")).Write(ws); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.Id, true, []byte{}).Write(ws); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", ws.RealRemote(), env.Event.Pubkey)
			ws.SetAuthed(string(env.Event.Pubkey))
		}
	}
	return
}
