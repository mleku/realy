package realy

import (
	"fmt"

	"realy.mleku.dev/context"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/closeenvelope"
	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/envelopes/noticeenvelope"
	"realy.mleku.dev/envelopes/reqenvelope"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/socketapi"
	"realy.mleku.dev/store"
)

func (s *Server) handleMessage(c context.T, a *socketapi.A, msg []byte, sto store.I) {
	var notice []byte
	var err error
	var t string
	var rem []byte
	if t, rem, err = envelopes.Identify(msg); chk.E(err) {
		notice = []byte(err.Error())
	}
	skipEventFunc := s.options.SkipEventFunc
	rl := s.relay
	switch t {
	case eventenvelope.L:
		notice = a.HandleEvent(c, rem, s)
	case reqenvelope.L:
		notice = a.HandleReq(c, rem, skipEventFunc, s)
	case closeenvelope.L:
		notice = a.HandleClose(rem, s)
	case authenvelope.L:
		notice = a.HandleAuth(rem, s)
	default:
		if wsh, ok := rl.(relay.WebSocketHandler); ok {
			wsh.HandleUnknownType(a.Listener, t, rem)
		} else {
			notice = []byte(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
		}
	}
	if len(notice) > 0 {
		log.D.F("notice->%s %s", a.RealRemote(), notice)
		if err = noticeenvelope.NewFrom(notice).Write(a.Listener); err != nil {
			return
		}
	}
}
