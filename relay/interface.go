package relay

import (
	"net/http"

	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/relayinfo"
	"realy.lol/store"
	"realy.lol/web"
)

// I is the main interface for implementing a nostr
type I interface {
	// Name is used as the "name" field in NIP-11 and as a prefix in default Server logging.
	// For other NIP-11 fields, see [Informationer].
	Name() st
	// Init is called at the very beginning by [Server.Start], allowing a realy
	// to initialize its internal resources.
	// Also see [eventstore.I.Init].
	Init() er
	// AcceptEvent is called for every nostr event received by the server.
	//
	// If the returned value is true, the event is passed on to [Storage.SaveEvent].
	// Otherwise, the server responds with a negative and "blocked" message as described
	// in NIP-20.
	//
	// Moderation via follow/mute lists of moderator npubs should deny events from
	// npubs listed in moderator mute lists. Events submitted by users not on the
	// moderator follow lists but submitting events containing p tags for direct
	// messages, that are not on the mute list, that do not yet have a reply, should accept
	// direct and group message events until there is three and thereafter will be restricted
	// until the user adds them to their follow list.
	AcceptEvent(c cx, ev *event.T, hr *http.Request, origin st,
		authedPubkey by) (accept bo,
		notice st, afterSave func())
	// Storage returns the realy storage implementation.
	Storage() store.I
}

// ReqAcceptor is the main interface for implementing a nostr
type ReqAcceptor interface {
	// AcceptReq is called for every nostr request filters received by the
	// server. If the returned value is true, the filters is passed on to
	// [Storage.QueryEvent].
	//
	// If moderation of access by follow/mute list of moderator npubs is enabled,
	// only users in the follow lists of mods are allowed read access (accepting
	// requests), all others should receive an OK,false,restricted response if
	// authed and if not authed CLOSED,restricted.
	//
	// If a user is not whitelisted by follow and not blacklisted by mute and the
	// request is for a message that contains their npub in a `p` tag that are
	// direct or group chat messages they also can be accepted, enabling full
	// support for in/outbox access.
	//
	// In order to support the ability to respond to
	AcceptReq(c cx, hr *http.Request, id by, ff *filters.T,
		authedPubkey by) (allowed *filters.T, ok bo)
}

// Authenticator is the interface for implementing NIP-42.
// ServiceURL() returns the URL used to verify the "AUTH" event from clients.
type Authenticator interface {
	AuthEnabled() bo
	ServiceUrl(r *http.Request) st
}

type Injector interface {
	InjectEvents() event.C
}

// Informationer is called to compose NIP-11 response to an HTTP request
// with application/nostr+json mime type.
// See also [I.Name].
type Informationer interface {
	GetNIP11InformationDocument() *relayinfo.T
}

// WebSocketHandler is passed nostr message types unrecognized by the
// server. The server handles "EVENT", "REQ" and "CLOSE" messages, as described in NIP-01.
type WebSocketHandler interface {
	HandleUnknownType(ws *web.Socket, t st, request by)
}

// ShutdownAware is called during the server shutdown.
// See [Server.Shutdown] for details.
type ShutdownAware interface {
	OnShutdown(cx)
}

// Logger is what [Server] uses to log messages.
type Logger interface {
	Infof(format st, v ...any)
	Warningf(format st, v ...any)
	Errorf(format st, v ...any)
}

// AdvancedDeleter methods are called before and after [Storage.DeleteEvent].
type AdvancedDeleter interface {
	BeforeDelete(ctx cx, id, pubkey by)
	AfterDelete(id, pubkey by)
}

// AdvancedSaver methods are called before and after [Storage.SaveEvent].
type AdvancedSaver interface {
	BeforeSave(cx, *event.T)
	AfterSave(*event.T)
}

type EventCounter interface {
	CountEvents(c cx, f *filter.T) (count no, approx bo, err er)
}
