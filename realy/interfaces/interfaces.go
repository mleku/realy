package interfaces

import (
	"net/http"
	"time"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/realy/options"
	"realy.mleku.dev/realy/publisher"
	"realy.mleku.dev/relay"
	"realy.mleku.dev/store"
)

type Server interface {
	AcceptEvent(
		c context.T, ev *event.T, hr *http.Request, origin string,
		authedPubkey []byte) (accept bool, notice string, afterSave func())
	AddEvent(
		c context.T, rl relay.I, ev *event.T, hr *http.Request,
		origin string, authedPubkey []byte) (accepted bool,
		message []byte)
	AdminAuth(r *http.Request,
		tolerance ...time.Duration) (authed bool, pubkey []byte)
	AuthRequired() bool
	Configuration() store.Configuration
	Context() context.T
	Disconnect()
	Publisher() *publisher.S
	Owners() [][]byte
	PublicReadable() bool
	Publish(c context.T, evt *event.T) (err error)
	Relay() relay.I
	SetConfiguration(*store.Configuration)
	Shutdown()
	Storage() store.I
	Options() *options.T
}
