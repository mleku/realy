package relay

import (
	"github.com/nbd-wtf/go-nostr/nip11"
	"mleku.dev/event"
	"mleku.dev/filter"
	"mleku.dev/filters"
	store "mleku.dev/store"
)

// Relay is the main interface for implementing a nostr relay.
type Relay interface {
	// Name is used as the "name" field in NIP-11 and as a prefix in default Server logging.
	// For other NIP-11 fields, see [Informationer].
	Name() S
	// Init is called at the very beginning by [Server.Start], allowing a relay
	// to initialize its internal resources.
	// Also see [eventstore.I.Init].
	Init() E
	// AcceptEvent is called for every nostr event received by the server.
	// If the returned value is true, the event is passed on to [Storage.SaveEvent].
	// Otherwise, the server responds with a negative and "blocked" message as described
	// in NIP-20.
	AcceptEvent(Ctx, *event.T) bool
	// Storage returns the relay storage implementation.
	Storage(Ctx) store.I
}

// ReqAcceptor is the main interface for implementing a nostr relay.
type ReqAcceptor interface {
	// AcceptReq is called for every nostr request filters received by the
	// server. If the returned value is true, the filtres is passed on to
	// [Storage.QueryEvent].
	AcceptReq(ctx Ctx, id B, ff *filters.T, authedPubkey B) bool
}

// Authenticator is the interface for implementing NIP-42.
// ServiceURL() returns the URL used to verify the "AUTH" event from clients.
type Authenticator interface {
	ServiceURL() S
}

type Injector interface {
	InjectEvents() event.C
}

// Informationer is called to compose NIP-11 response to an HTTP request
// with application/nostr+json mime type.
// See also [Relay.Name].
type Informationer interface {
	GetNIP11InformationDocument() nip11.RelayInformationDocument
}

// CustomWebSocketHandler is passed nostr message types unrecognized by the
// server. The server handles "EVENT", "REQ" and "CLOSE" messages, as described in NIP-01.
type CustomWebSocketHandler interface {
	HandleUnknownType(ws *WebSocket, t S, request B)
}

// ShutdownAware is called during the server shutdown.
// See [Server.Shutdown] for details.
type ShutdownAware interface {
	OnShutdown(Ctx)
}

// Logger is what [Server] uses to log messages.
type Logger interface {
	Infof(format S, v ...any)
	Warningf(format S, v ...any)
	Errorf(format S, v ...any)
}

// AdvancedDeleter methods are called before and after [Storage.DeleteEvent].
type AdvancedDeleter interface {
	BeforeDelete(ctx Ctx, id, pubkey B)
	AfterDelete(id, pubkey B)
}

// AdvancedSaver methods are called before and after [Storage.SaveEvent].
type AdvancedSaver interface {
	BeforeSave(Ctx, *event.T)
	AfterSave(*event.T)
}

type EventCounter interface {
	CountEvents(c Ctx, f *filter.T) (int64, E)
}