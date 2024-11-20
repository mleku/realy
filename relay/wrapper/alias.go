package wrapper

import (
	"net/http"

	"realy.lol/ws"

	"realy.lol/envelopes/okenvelope"
	"realy.lol/subscription"
)

type SubID = subscription.Id
type WS = *ws.Serv
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
