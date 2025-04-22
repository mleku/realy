package socketapi

import (
	"fmt"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/closeenvelope"
	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/envelopes/noticeenvelope"
	"realy.mleku.dev/envelopes/reqenvelope"
	"realy.mleku.dev/log"
	"realy.mleku.dev/relay"
)

func (a *A) HandleMessage(msg []byte) {
	var notice []byte
	var err error
	var t string
	var rem []byte
	if t, rem = envelopes.Identify(msg); chk.E(err) {
		notice = []byte(err.Error())
	}
	rl := a.Relay()
	switch t {
	case eventenvelope.L:
		notice = a.HandleEvent(a.Context(), rem, a.Server)
	case reqenvelope.L:
		notice = a.HandleReq(a.Context(), rem, a.Options().SkipEventFunc, a.Server)
	case closeenvelope.L:
		notice = a.HandleClose(rem, a.Server)
	case authenvelope.L:
		notice = a.HandleAuth(rem, a.Server)
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
