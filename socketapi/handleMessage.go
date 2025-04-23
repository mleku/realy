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
)

func (a *A) HandleMessage(msg []byte, remote string) {
	log.T.F("received message from %s\n%s", remote, msg)
	var notice []byte
	var err error
	var t string
	var rem []byte
	if t, rem = envelopes.Identify(msg); chk.E(err) {
		notice = []byte(err.Error())
	}
	switch t {
	case eventenvelope.L:
		notice = a.HandleEvent(a.Context(), rem, a.Server, remote)
	case reqenvelope.L:
		notice = a.HandleReq(a.Context(), rem, a.Server, remote)
	case closeenvelope.L:
		notice = a.HandleClose(rem, a.Server)
	case authenvelope.L:
		notice = a.HandleAuth(rem, a.Server)
	default:
		notice = []byte(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
	}
	if len(notice) > 0 {
		log.D.F("notice->%s %s", remote, notice)
		if err = noticeenvelope.NewFrom(notice).Write(a.Listener); err != nil {
			return
		}
	}

}
