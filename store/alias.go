package store

import (
	"net/http"

	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/subscription"
)

type SubID = subscription.Id
type Responder = http.ResponseWriter
type Req = *http.Request
type OK = okenvelope.T
