package store

import (
	"net/http"

	"realy.lol/envelopes/okenvelope"
	"realy.lol/subscription"
)

type SubID = subscription.Id
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
