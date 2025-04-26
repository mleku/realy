// Package relay contains a collection of interfaces for enabling the building
// of modular nostr relay implementations.
package relay

import (
	"net/http"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/store"
)

// I is the main interface for implementing a nostr relay.
type I interface {
	// Name is used as the "name" field in NIP-11 and as a prefix in default Server logging.
	// For other NIP-11 fields, see [Informationer].
	Name() string
	// Init is called at the very beginning by [Server.Start], allowing a realy
	// to initialize its internal resources.
	// Also see [eventstore.I.Init].
	Init() error
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
	AcceptEvent(c context.T, ev *event.T, hr *http.Request, origin string,
		authedPubkey []byte) (accept bool, notice string, afterSave func())
	// Storage returns the realy storage implementation.
	Storage() store.I
	// Owners returns the list of pubkeys designated as owners of the relay.
	Owners() [][]byte
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
	AcceptReq(c context.T, hr *http.Request, id []byte, ff *filters.T,
		authedPubkey []byte) (allowed *filters.T,
		ok bool, modified bool)
	// AcceptFilter is basically the same as AcceptReq except it is additional to
	// enable the simplified filter query type.
	AcceptFilter(c context.T, hr *http.Request, f *filter.S,
		authedPubkey []byte) (allowed *filter.S, ok bool, modified bool)
	AuthRequired() bool
	ServiceUrl(r *http.Request) string
	CheckOwnerLists(c context.T)
	ZeroLists()
	AllFollowed(pk []byte) (ok bool)
	OwnersFollowed(pk []byte) (ok bool)
	Muted(pk []byte) (ok bool)
}
