package socketapi

import (
	"realy.mleku.dev/auth"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/normalize"
	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/relay"
)

func (a *A) HandleAuth(req []byte,
	srv interfaces.Server) (msg []byte) {

	if auther, ok := srv.Relay().(relay.Authenticator); ok && auther.AuthRequired() {
		svcUrl := auther.ServiceUrl(a.Req())
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
		if valid, err = auth.Validate(env.Event, []byte(a.Challenge()),
			svcUrl); chk.E(err) {
			e := err.Error()
			if err = okenvelope.NewFrom(env.Event.Id, false,
				normalize.Error.F(err.Error())).Write(a.Listener); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Error.F(e)
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.Id, false,
				normalize.Error.F("failed to authenticate")).Write(a.Listener); chk.E(err) {
				return []byte(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.Id, true,
				[]byte{}).Write(a.Listener); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", a.RealRemote(), env.Event.Pubkey)
			a.SetAuthed(string(env.Event.Pubkey))
		}
	}
	return
}
