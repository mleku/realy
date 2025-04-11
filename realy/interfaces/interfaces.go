package interfaces

import (
	"net/http"
	"time"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/realy/subscribers"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/store"
)

type Server interface {
	Context() context.T
	AdminAuth(r *http.Request,
		tolerance ...time.Duration) (authed bool, pubkey []byte)
	Storage() store.I
	Configuration() store.Configuration
	SetConfiguration(*store.Configuration)
	Relay() relay.I
	Disconnect()
	AddEvent(
		c context.T, rl relay.I, ev *event.T, hr *http.Request,
		origin string, authedPubkey []byte) (accepted bool,
		message []byte)
	AcceptEvent(
		c context.T, ev *event.T, hr *http.Request, origin string,
		authedPubkey []byte) (accept bool, notice string, afterSave func())
	Listeners() *subscribers.S
	PublicReadable() bool
	Owners() [][]byte
	Shutdown()
}
