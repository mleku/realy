package eventstore

import (
	"net/http"

	"realy.lol/ws"

	"realy.lol/envelopes/okenvelope"
	"realy.lol/subscriptionid"
)

type SubID = subscriptionid.T
type WS = *ws.Serv
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
