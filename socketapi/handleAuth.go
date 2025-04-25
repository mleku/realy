package socketapi

import (
	"realy.mleku.dev/auth"
	"realy.mleku.dev/chk"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/reason"
)

func (a *A) HandleAuth(b []byte, srv interfaces.Server) (msg []byte) {
	if srv.AuthRequired() || len(a.Owners()) > 0 || !a.PublicReadable() {
		svcUrl := srv.ServiceURL(a.Listener.Req())
		if svcUrl == "" {
			return
		}
		log.T.F("received auth response,%s", b)
		var err error
		var rem []byte
		env := authenvelope.NewResponse()
		if rem, err = env.Unmarshal(b); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.F("extra '%s'", rem)
		}
		var valid bool
		if valid, err = auth.Validate(env.Event, []byte(a.Listener.Challenge()),
			svcUrl); chk.E(err) {
			e := err.Error()
			if err = Ok.Error(a, env, e); chk.E(err) {
				return []byte(e)
			}
			return reason.Error.F(e)
		} else if !valid {
			if err = Ok.Error(a, env, "failed to authenticate"); chk.E(err) {
				return
			}
			return reason.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.Id, true).Write(a.Listener); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", a.Listener.RealRemote(), env.Event.Pubkey)
			a.Listener.SetAuthed(string(env.Event.Pubkey))
		}
	}
	return
}
