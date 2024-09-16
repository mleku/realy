package eventstore

import (
	"net/http"

	"mleku.dev/ws"

	"mleku.dev/envelopes/okenvelope"
	"mleku.dev/subscriptionid"
)

type SubID = subscriptionid.T
type WS = *ws.Serv
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
