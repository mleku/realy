package app

import (
	"net/http"
	"strconv"
	"strings"

	"realy.lol/context"
	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/filters"
	"realy.lol/relay"
	"realy.lol/store"
)

type Relay struct {
	*Config
	Store store.I
}

func (r *Relay) Name() S                     { return "REALY" }
func (r *Relay) Storage(c context.T) store.I { return r.Store }
func (r *Relay) Init() E                     { return nil }
func (r *Relay) AcceptEvent(c context.T, evt *event.T) bool {
	// c.Value()
	return true
}

func (r *Relay) AcceptReq(c Ctx, id B, ff *filters.T, authedPubkey B) bool {
	// if the authenticator is enabled we require auth to process requests
	if _, ok := (relay.I)(r).(relay.Authenticator); ok {
		return len(authedPubkey) == schnorr.PubKeyBytesLen
	}
	return true
}

// ServiceUrl returns the address of the relay to send back in auth responses.
// If this is implemented it enables auth-required.
func (r *Relay) ServiceUrl(req *http.Request) (s S) {
	log.I.S(req)
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
