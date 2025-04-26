package socketapi

import (
	"realy.lol/chk"
	"realy.lol/envelopes/closeenvelope"
	"realy.lol/log"
	"realy.lol/publish"
	"realy.lol/realy/interfaces"
)

func (a *A) HandleClose(req []byte,
	srv interfaces.Server) (note []byte) {
	var err error
	var rem []byte
	env := closeenvelope.New()
	if rem, err = env.Unmarshal(req); chk.E(err) {
		return []byte(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.ID.String() == "" {
		return []byte("CLOSE has no <id>")
	}
	publish.P.Receive(&W{
		Cancel:   true,
		Listener: a.Listener,
		Id:       env.ID.String(),
	})
	// srv.Publisher().removeSubscriberId(a.Listener, env.ID.String())
	return
}
