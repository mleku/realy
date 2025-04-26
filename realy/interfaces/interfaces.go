package interfaces

import (
	"net/http"
	"time"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filters"
	"realy.lol/realy/config"
	"realy.lol/store"
)

type Server interface {
	AcceptEvent(c context.T, ev *event.T, hr *http.Request, authedPubkey []byte, remote string) (accept bool, notice string, afterSave func())
	AcceptReq(c context.T, hr *http.Request, id []byte, f *filters.T, authedPubkey []byte, remote string) (allowed *filters.T, ok bool, modified bool)
	AddEvent(c context.T, ev *event.T, hr *http.Request, authedPubkey []byte, remote string) (accepted bool, message []byte)
	AdminAuth(r *http.Request, remote string, tolerance ...time.Duration) (authed bool, pubkey []byte)
	AuthRequired() bool
	CheckOwnerLists(c context.T)
	Configuration() config.C
	Context() context.T
	HandleRelayInfo(w http.ResponseWriter, r *http.Request)
	Lock()
	Owners() [][]byte
	OwnersFollowed(pubkey string) (ok bool)
	PublicReadable() bool
	ServiceURL(req *http.Request) (s string)
	SetConfiguration(cfg config.C) (err error)
	Shutdown()
	Storage() store.I
	Unlock()
	UpdateConfiguration() (err error)
	ZeroLists()
}
