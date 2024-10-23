package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"realy.lol/context"
	"realy.lol/ec/schnorr"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/store"
	"realy.lol/tag"
)

type Relay struct {
	*Config
	Store           store.I
	Owners          []B
	Followed, Muted map[S]struct{}
	sync.Mutex
}

func (r *Relay) Name() S                     { return "REALY" }
func (r *Relay) Storage(c context.T) store.I { return r.Store }
func (r *Relay) Init() (err E) {
	for _, src := range r.Config.Owners {
		dst := make(B, len(src)/2)
		if _, err = hex.DecBytes(dst, B(src)); chk.E(err) {
			continue
		}
		r.Owners = append(r.Owners, dst)
	}
	r.CheckOwnerLists(context.Bg())
	return nil
}
func (r *Relay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, authedPubkey B) bool {
	// if the authenticator is enabled we require auth to accept events
	if !r.AuthEnabled() {
		log.I.F("auth not enabled")
		return true
	}
	if len(authedPubkey) != 32 {
		log.E.F("client not authed with auth required")
		return false
	}
	if len(r.Owners) > 0 {
		r.Lock()
		defer r.Unlock()
		if evt.Kind.Equal(kind.FollowList) || evt.Kind.Equal(kind.MuteList) {
			for _, o := range r.Owners {
				log.I.F("own %0x\npub %0x", o, evt.PubKey)
				if equals(o, evt.PubKey) {
					// owner has updated follows or mute list, so we zero those lists so they
					// are regenerated for the next AcceptReq/AcceptEvent
					r.Followed = make(map[S]struct{})
					r.Muted = make(map[S]struct{})
					log.I.F("clearing owner follow/mute lists because of update from %0x",
						evt.PubKey)
					return true
				}
			}
		}
		for _, o := range r.Owners {
			log.T.F("%0x,%0x", o, evt.PubKey)
			if equals(o, evt.PubKey) {
				log.W.Ln("event is from owner")
				return true
			}
		}
		// check the mute list, and reject events authored by muted pubkeys, even if
		// they come from a pubkey that is on the follow list.
		for pk := range r.Muted {
			if equals(evt.PubKey, B(pk)) {
				log.I.F("rejecting event with pubkey %v because on owner mute list",
					evt.PubKey)
				return false
			}
		}
		// for all else, check the authed pubkey is in the follow list
		for pk := range r.Followed {
			if equals(authedPubkey, B(pk)) {
				log.I.F("accepting event %0x because on owner follow list", evt.ID)
				return true
			}
		}
		// if the authed pubkey was not found, reject the request.
		// log.I.F("authed pubkey %0x not found, rejecting event", authedPubkey)
		// return false
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
	// if the client hasn't authed, reject
	if len(authedPubkey) == 0 {
		return false
	}
	// regenerate lists if they have been updated
	r.CheckOwnerLists(c)
	// check that the client is authed to a pubkey in the owner follow list
	if len(r.Owners) > 0 {
		for pk := range r.Followed {
			if equals(authedPubkey, B(pk)) {
				return true
			}
		}
		// if the authed pubkey was not found, reject the request.
		return false
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	return len(authedPubkey) == schnorr.PubKeyBytesLen
}

// CheckOwnerLists regenerates the owner follow and mute lists if they are empty
func (r *Relay) CheckOwnerLists(c context.T) {
	if len(r.Owners) > 0 {
		r.Lock()
		defer r.Unlock()
		// need to search DB for moderator npub follow lists, followed npubs are allowed access.
		if len(r.Followed) < 1 {
			log.I.Ln("regenerating owners follow lists")
			var err error
			var evs []*event.T
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.Owners...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {

			}
			// preallocate sufficient elements
			var count int
			for _, ev := range evs {
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B{'p'}) {
						count++
					}
				}
			}
			r.Followed = make(map[S]struct{})
			for _, ev := range evs {
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B{'p'}) {
						var dst B
						if dst, err = hex.DecAppend(dst, t.Value()); chk.E(err) {
							continue
						}
						r.Followed[S(dst)] = struct{}{}
					}
				}
			}
		}
		if len(r.Muted) < 1 {
			log.I.Ln("regenerating owners mute lists")
			var err error
			var evs []*event.T
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.Owners...),
					Kinds: kinds.New(kind.MuteList)}); chk.E(err) {

			}
			// // preallocate sufficient elements
			// var count int
			// for _, ev := range evs {
			// 	for _, t := range ev.Tags.F() {
			// 		if equals(t.Key(), B{'p'}) {
			// 			count++
			// 		}
			// 	}
			// }
			r.Muted = make(map[S]struct{})
			for _, ev := range evs {
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B{'p'}) {
						var dst B
						if dst, err = hex.DecAppend(dst, t.Value()); chk.E(err) {
							continue
						}
						r.Muted[S(dst)] = struct{}{}
					}
				}
			}
			o := "followed:\n"
			for pk := range r.Followed {
				o += fmt.Sprintf("%x,", pk)
			}
			o += "\nmuted:\n"
			for pk := range r.Muted {
				o += fmt.Sprintf("%x,", pk)
			}
			// log.T.F("%s\n", o)
		}
	}
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
