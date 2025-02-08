package realy

import (
	"realy.lol/envelopes/closeenvelope"
	"realy.lol/web"
)

func (s *Server) handleClose(ws *web.Socket, req []byte) (note []byte) {
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
	s.listeners.RemoveListenerId(ws, env.ID.String())
	return
}
