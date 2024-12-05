package realy

import (
	"realy.lol/envelopes/closeenvelope"
	"realy.lol/web"
)

func (s *Server) handleClose(ws *web.Socket, req by) (note by) {
	var err er
	var rem by
	env := closeenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return by(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.ID.String() == "" {
		return by("CLOSE has no <id>")
	}
	s.listeners.RemoveListenerId(ws, env.ID.String())
	return
}
