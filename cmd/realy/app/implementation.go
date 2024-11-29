package app

import (
	"bytes"
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
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/store"
	"realy.lol/tag"
)

type Relay struct {
	sync.Mutex
	*Config
	Store store.I
	// Owners' pubkeys
	Owners          []B
	Followed, Muted map[S]struct{}
	// OwnersFollowLists are the event IDs of owners follow lists, which must not be deleted, only
	// replaced.
	OwnersFollowLists []B
	// OwnersMuteLists are the event IDs of owners mute lists, which must not be deleted, only
	// replaced.
	OwnersMuteLists []B
}

func (r *Relay) Name() S { return r.Config.AppName }

func (r *Relay) Storage(c context.T) store.I { return r.Store }

func (r *Relay) Init() (err E) {
	for _, src := range r.Config.Owners {
		if len(src) < 1 {
			continue
		}
		dst := make(B, len(src)/2)
		if _, err = hex.DecBytes(dst, B(src)); chk.E(err) {
			continue
		}
		r.Owners = append(r.Owners, dst)
	}
	log.T.C(func() string {
		ownerIds := make([]string, len(r.Owners))
		for i, npub := range r.Owners {
			ownerIds[i] = hex.Enc(npub)
		}
		return fmt.Sprintf("%v", ownerIds)
	})
	r.Followed = make(map[S]struct{})
	r.Muted = make(map[S]struct{})
	r.CheckOwnerLists(context.Bg())
	return nil
}

func (r *Relay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, origin S,
	authedPubkey B) (accept bool, notice S) {
	// if the authenticator is enabled we require auth to accept events
	if !r.AuthEnabled() {
		return true, ""
	}
	if len(authedPubkey) != 32 {
		return false, fmt.Sprintf("client not authed with auth required %s", origin)
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
					r.OwnersFollowLists = r.OwnersFollowLists[:0]
					r.Muted = make(map[S]struct{})
					r.OwnersMuteLists = r.OwnersMuteLists[:0]
					log.I.F("clearing owner follow/mute lists because of update from %s %0x",
						origin, evt.PubKey)
					return true, ""
				}
			}
		}
		for _, o := range r.Owners {
			log.T.F("%0x,%0x", o, evt.PubKey)
			if equals(o, evt.PubKey) {
				// prevent owners from deleting their own mute/follow lists in case of bad
				// client implementation
				if evt.Kind.Equal(kind.Deletion) {
					// we don't accept deletes on owners' follow or mute lists because of the
					// potential for a malicious action causing this, first check for the list:
					tt := tag.New(append(r.OwnersFollowLists, r.OwnersMuteLists...)...)
					if evt.Tags.ContainsAny(B("e"), tt) {
						return false, "cannot delete owner's follow, owners's follows follow or mute events"
					}
					// next, check all a tags present are not follow/mute lists of the owners
					aTags := evt.Tags.GetAll(tag.New("a"))
					for _, at := range aTags.F() {
						split := bytes.Split(at.Value(), B{':'})
						if len(split) != 3 {
							continue
						}
						kin := ints.New(uint16(0))
						if _, err := kin.UnmarshalJSON(split[0]); chk.E(err) {
							return
						}
						kk := kind.New(kin.Uint16())
						if kk.Equal(kind.Deletion) {
							// we don't delete delete events, period
							return false, "delete event kind may not be deleted"
						}
						// if the kind is not parameterised replaceable, the tag is invalid and the
						// delete event will not be saved.
						if !kk.IsParameterizedReplaceable() {
							return false, "delete tags with a tags containing " +
								"non-parameterized-replaceable events cannot be processed"
						}
						for _, own := range r.Owners {
							// don't allow owners to delete their mute or follow lists because
							// they should not want to, can simply replace it, and malicious
							// clients may do this specifically to attack the owner's relay (s)
							if equals(own, split[1]) ||
								kk.Equal(kind.MuteList) ||
								kk.Equal(kind.FollowList) {
								return false, "owners may not delete their own " +
									"mute or follow lists, they can be replaced"
							}
						}
					}
					log.W.Ln("event is from owner")
					return true, ""
				}
			}
			// check the mute list, and reject events authored by muted pubkeys, even if
			// they come from a pubkey that is on the follow list.
			for pk := range r.Muted {
				if equals(evt.PubKey, B(pk)) {
					return false, "rejecting event with pubkey " + S(evt.PubKey) +
						" because on owner mute list"
				}
			}
			// for all else, check the authed pubkey is in the follow list
			for pk := range r.Followed {
				// allow all events from follows of owners
				if equals(authedPubkey, B(pk)) {
					log.I.F("accepting event %0x because %0x on owner follow list",
						evt.ID, B(pk))
					return true, ""
				}
			}
		}
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	accept = len(authedPubkey) == schnorr.PubKeyBytesLen
	if !accept {
		notice = "auth required but user not authed"
	}
	return
}

func (r *Relay) AcceptReq(c Ctx, hr *http.Request, idB, ff *filters.T,
	authedPubkey B) (allowed *filters.T, ok bool) {
	// if the authenticator is enabled we require auth to process requests
	if !r.AuthEnabled() {
		ok = true
		return
	}
	// if client isn't authed but there are kinds in the filters that are
	// kind.Directory type then trim the filter down and only respond to the queries
	// that blanket should deliver events in order to facilitate non-authorized users
	// to interact with users, even just such as to see their profile metadata or
	// learn about deleted events.
	if len(authedPubkey) == 0 {
		for _, f := range ff.F {
			fk := f.Kinds.K
			allowedKinds := kinds.New()
			for _, fkk := range fk {
				if fkk.IsDirectoryEvent() {
					allowedKinds.K = append(allowedKinds.K, fkk)
				}
			}
			// if none of the kinds in the req are permitted, continue to the next filter.
			if len(allowedKinds.K) == 0 {
				continue
			}
			// if no filters have yet been added, initialize one
			if allowed == nil {
				allowed = &filters.T{}
			}
			// overwrite the kinds that have been permitted
			f.Kinds.K = allowedKinds.K
			allowed.F = append(allowed.F, f)
		}
		if allowed != nil {
			// request has been filtered and can be processed. note that the caller should
			// still send out an auth request after the filter has been processed.
			ok = true
			return
		}
	}
	// if the client hasn't authed, reject
	if len(authedPubkey) == 0 {
		return
	}
	// client is permitted, pass through the filter so request/count processing does
	// not need logic and can just use the returned filter.
	allowed = ff
	// regenerate lists if they have been updated
	r.CheckOwnerLists(c)
	// check that the client is authed to a pubkey in the owner follow list
	if len(r.Owners) > 0 {
		for pk := range r.Followed {
			if equals(authedPubkey, B(pk)) {
				ok = true
				return
			}
		}
		// if the authed pubkey was not found, reject the request.
		return
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	ok = len(authedPubkey) == schnorr.PubKeyBytesLen
	return
}

// CheckOwnerLists regenerates the owner follow and mute lists if they are empty.
//
// It also adds the followed npubs of the follows.
func (r *Relay) CheckOwnerLists(c context.T) {
	if len(r.Owners) > 0 {
		r.Lock()
		defer r.Unlock()
		var err error
		var evs []*event.T
		// need to search DB for moderator npub follow lists, followed npubs are allowed access.
		if len(r.Followed) < 1 {
			log.D.Ln("regenerating owners follow lists")
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.Owners...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				r.OwnersFollowLists = append(r.OwnersFollowLists, ev.ID)
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B("p")) {
						var p B
						if p, err = hex.Dec(S(t.Value())); chk.E(err) {
							continue
						}
						r.Followed[S(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
			// next, search for the follow lists of all on the follow list
			log.D.Ln("searching for owners follows follow lists")
			var followed []S
			for f := range r.Followed {
				followed = append(followed, f)
			}
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(followed...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				// we want to protect the follow lists of users as well so they also cannot be
				// deleted, only replaced.
				r.OwnersFollowLists = append(r.OwnersFollowLists, ev.ID)
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B("p")) {
						var p B
						if p, err = hex.Dec(S(t.Value())); chk.E(err) {
							continue
						}
						r.Followed[S(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		if len(r.Muted) < 1 {
			log.D.Ln("regenerating owners mute lists")
			r.Muted = make(map[S]struct{})
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.Owners...),
					Kinds: kinds.New(kind.MuteList)}); chk.E(err) {
			}
			for _, ev := range evs {
				r.OwnersMuteLists = append(r.OwnersMuteLists, ev.ID)
				for _, t := range ev.Tags.F() {
					if equals(t.Key(), B("p")) {
						var p B
						if p, err = hex.Dec(S(t.Value())); chk.E(err) {
							continue
						}
						r.Muted[S(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		// log this info
		o := "followed:\n"
		for pk := range r.Followed {
			o += fmt.Sprintf("%0x,", pk)
		}
		o += "\nmuted:\n"
		for pk := range r.Muted {
			o += fmt.Sprintf("%0x,", pk)
		}
		log.T.F("%s\n", o)
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
