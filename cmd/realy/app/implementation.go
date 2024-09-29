package app

import (
	"net/http"
	"strconv"
	"strings"

	"realy.lol/context"
	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/filters"
	"realy.lol/store"
)

type Relay struct {
	*Config
	Store store.I
}

func (r *Relay) Name() S                     { return "REALY" }
func (r *Relay) Storage(c context.T) store.I { return r.Store }
func (r *Relay) Init() E                     { return nil }
func (r *Relay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, authedPubkey B) bool {
	// if the authenticator is enabled we require auth to accept events
	if !r.AuthEnabled() {
		return true
	}
	// check for moderator npubs follow/mute lists if they exist, and accept based on this
	if len(r.Moderators) > 0 {
		// need to search DB for moderator npub follow and mute lists and allow access
		// accordingly
		return true
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	return len(authedPubkey) == schnorr.PubKeyBytesLen
}

func (r *Relay) AcceptReq(c Ctx, hr *http.Request, id B, ff *filters.T, authedPubkey B) bool {
	// if the authenticator is enabled we require auth to process requests
	if !r.AuthEnabled() {
		return true
	}
	// check for moderator npubs follow/mute lists if they exist, and accept based on this
	if len(r.Moderators) > 0 {
		// need to search DB for moderator npub follow and mute lists and allow access
		// accordingly
		return true
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	return len(authedPubkey) == schnorr.PubKeyBytesLen
}

func (r *Relay) AuthEnabled() bool { return r.Config.AuthRequired }

// ServiceUrl returns the address of the relay to send back in auth responses.
// If auth is disabled this returns an empty string.
func (r *Relay) ServiceUrl(req *http.Request) (s S) {
	if !r.Config.AuthRequired {
		return
	}
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}
	proto := req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if host == "localhost" {
			proto = "ws"
		} else if strings.Contains(host, ":") {
			// has a port number
			proto = "ws"
		} else if _, err := strconv.Atoi(strings.ReplaceAll(host, ".",
			"")); chk.E(err) {
			// it's a naked IP
			proto = "ws"
		} else {
			proto = "wss"
		}
	} else if proto == "https" {
		proto = "wss"
	} else if proto == "http" {
		proto = "ws"
	}
	return proto + "://" + host
}
