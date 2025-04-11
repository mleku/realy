package realy

import (
	"realy.mleku.dev/envelopes/closeenvelope"
	"realy.mleku.dev/ws"
)

func (s *Server) handleClose(ws *ws.Listener, req []byte) (note []byte) {
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
	s.listeners.RemoveSubscriberId(ws, env.ID.String())
	return
}
