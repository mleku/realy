// Package app implements the realy nostr relay with a simple follow/mute list authentication scheme and the new HTTP REST based protocol.
package app

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"realy.mleku.dev/bech32encoding"
	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/ec/schnorr"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/kinds"
	"realy.mleku.dev/list"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/config"
	"realy.mleku.dev/store"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/tag/atag"
)

type Relay struct {
	sync.Mutex
	*config.C
	Store store.I
	// Owners' pubkeys
	owners [][]byte
	// allFollowed are the pubkeys that are in the Owners' follow lists and have full
	// access permission.
	allFollowed list.L
	// OwnersFollowed are "guests" of the allFollowed and have full access but with
	// rate limiting enabled.
	ownersFollowed list.L
	// Muted are on Owners' mute lists and do not have write access to the relay,
	// even if they would be in the OwnersFollowed list, they can only read.
	muted list.L
	// ownersFollowLists are the event IDs of owners follow lists, which must not be
	// deleted, only replaced.
	ownersFollowLists [][]byte
	// ownersMuteLists are the event IDs of owners mute lists, which must not be
	// deleted, only replaced.
	ownersMuteLists [][]byte
}

func (r *Relay) Name() string { return r.C.AppName }

func (r *Relay) Storage() store.I { return r.Store }

func (r *Relay) Init() (err error) {
	for _, src := range r.C.Owners {
		if len(src) < 1 {
			continue
		}
		dst := make([]byte, len(src)/2)
		if _, err = hex.DecBytes(dst, []byte(src)); chk.E(err) {
			if dst, err = bech32encoding.NpubToBytes([]byte(src)); chk.E(err) {
				continue
			}
		}
		r.owners = append(r.owners, dst)
	}
	if len(r.owners) > 0 {
		log.T.C(func() string {
			ownerIds := make([]string, len(r.owners))
			for i, npub := range r.owners {
				ownerIds[i] = hex.Enc(npub)
			}
			owners := strings.Join(ownerIds, ",")
			return fmt.Sprintf("owners %s", owners)
		})
		r.ZeroLists()
		r.CheckOwnerLists(context.Bg())
	}
	return nil
}

func (r *Relay) Owners() [][]byte {
	r.Lock()
	defer r.Unlock()
	return r.owners
}

func (r *Relay) OwnersFollowed(pk []byte) (ok bool) {
	r.Lock()
	defer r.Unlock()
	_, ok = r.ownersFollowed[string(pk)]
	return
}
func (r *Relay) AllFollowed(pk []byte) (ok bool) {
	r.Lock()
	defer r.Unlock()
	_, ok = r.allFollowed[string(pk)]
	return
}

func (r *Relay) Muted(pk []byte) (ok bool) {
	r.Lock()
	defer r.Unlock()
	_, ok = r.muted[string(pk)]
	return
}

func (r *Relay) ZeroLists() {
	r.Lock()
	defer r.Unlock()
	r.allFollowed = make(map[string]struct{})
	r.ownersFollowed = make(map[string]struct{})
	r.ownersFollowLists = r.ownersFollowLists[:0]
	r.muted = make(map[string]struct{})
	r.ownersMuteLists = r.ownersMuteLists[:0]
}

func (r *Relay) AcceptEvent(c context.T, evt *event.T, hr *http.Request,
	origin string, authedPubkey []byte) (accept bool, notice string, afterSave func()) {
	// if the authenticator is enabled we require auth to accept events
	if !r.AuthRequired() && len(r.owners) < 1 {
		return true, "", nil
	}
	// if evt.CreatedAt.I64()-10 > time.Now().Unix() {
	// 	return false,
	// 		"realy does not accept timestamps that are so obviously fake, fix your clock",
	// 		nil
	// }
	if len(authedPubkey) != 32 && !r.PublicReadable {
		return false, fmt.Sprintf("client not authed with auth required %s", origin), nil
	}
	if len(r.owners) > 0 {
		r.Lock()
		defer r.Unlock()
		if evt.Kind.Equal(kind.FollowList) {
			// if owner or any of their follows lists are updated we need to regenerate the
			// list this ensures that immediately a follow changes their list that newly
			// allFollowed can access the relay and upload DM events and such for owner
			// allFollowed users.
			for o := range r.ownersFollowed {
				if bytes.Equal([]byte(o), evt.Pubkey) {
					return true, "", func() {
						r.ZeroLists()
						r.CheckOwnerLists(context.Bg())
					}
				}
			}
		}
		if evt.Kind.Equal(kind.MuteList) {
			// only owners control the mute list
			for _, o := range r.owners {
				if bytes.Equal(o, evt.Pubkey) {
					return true, "", func() {
						r.ZeroLists()
						r.CheckOwnerLists(context.Bg())
					}
				}
			}
		}
		for _, o := range r.owners {
			log.T.F("%0x,%0x", o, evt.Pubkey)
			if bytes.Equal(o, evt.Pubkey) {
				// prevent owners from deleting their own mute/follow lists in case of bad
				// client implementation
				if evt.Kind.Equal(kind.Deletion) {
					// check all a tags present are not follow/mute lists of the owners
					aTags := evt.Tags.GetAll(tag.New("a"))
					for _, at := range aTags.ToSliceOfTags() {
						a := &atag.T{}
						var rem []byte
						var err error
						if rem, err = a.Unmarshal(at.Value()); chk.E(err) {
							continue
						}
						if len(rem) > 0 {
							log.I.S("remainder", evt, rem)
						}
						if a.Kind.Equal(kind.Deletion) {
							// we don't delete delete events, period
							return false, "delete event kind may not be deleted", nil
						}
						// if the kind is not parameterised replaceable, the tag is invalid and the
						// delete event will not be saved.
						if !a.Kind.IsParameterizedReplaceable() {
							return false, "delete tags with a tags containing " +
								"non-parameterized-replaceable events cannot be processed", nil
						}
						for _, own := range r.owners {
							// don't allow owners to delete their mute or follow lists because
							// they should not want to, can simply replace it, and malicious
							// clients may do this specifically to attack the owner's relay (s)
							if bytes.Equal(own, a.PubKey) ||
								a.Kind.Equal(kind.MuteList) ||
								a.Kind.Equal(kind.FollowList) {
								return false, "owners may not delete their own " +
									"mute or follow lists, they can be replaced", nil
							}
						}
					}
					log.W.Ln("event is from owner")
					return true, "", nil
				}
			}
			// check the mute list, and reject events authored by muted pubkeys, even if
			// they come from a pubkey that is on the follow list.
			for pk := range r.muted {
				if bytes.Equal(evt.Pubkey, []byte(pk)) {
					return false, "rejecting event with pubkey " + hex.Enc(evt.Pubkey) +
						" because on owner mute list", nil
				}
			}
			// for all else, check the authed pubkey is in the follow list
			for pk := range r.allFollowed {
				// allow all events from follows of owners
				if bytes.Equal(authedPubkey, []byte(pk)) {
					log.I.F("accepting event %0x because %0x on owner follow list",
						evt.Id, []byte(pk))
					return true, "", nil
				}
			}
		}
	}
	// if auth is enabled and there is no moderators we just check that the pubkey
	// has been loaded via the auth function.
	accept = len(authedPubkey) == schnorr.PubKeyBytesLen
	if !accept {
		notice = "auth required but user not authed"
		afterSave = func() {

		}
	}
	return
}

func (r *Relay) AcceptFilter(c context.T, hr *http.Request, f *filter.S,
	authedPubkey []byte) (allowed *filter.S, ok bool, modified bool) {
	if r.PublicReadable && len(r.owners) == 0 {
		allowed = f
		ok = true
		return
	}
	// if client isn't authed but there are kinds in the filters that are
	// kind.Directory type then trim the filter down and only respond to the queries
	// that blanket should deliver events in order to facilitate non-authorized users
	// to interact with users, even just such as to see their profile metadata or
	// learn about deleted events.
	if len(authedPubkey) == 0 {
		fk := f.Kinds.K
		allowedKinds := kinds.New()
		for _, fkk := range fk {
			if fkk.IsDirectoryEvent() || (!fkk.IsPrivileged() && r.PublicReadable) {
				allowedKinds.K = append(allowedKinds.K, fkk)
			}
		}
		if len(fk) > 0 && len(allowedKinds.K) == 0 {
			// the filter has kinds, and none were permitted, this filter cannot be
			// processed.
			return
		}
		// overwrite the kinds that have been permitted
		if len(f.Kinds.K) != len(allowedKinds.K) {
			modified = true
		}
		f.Kinds.K = allowedKinds.K
		// we can process what remains
		ok = true
		return
	}
	// if the client hasn't authed, reject
	if len(authedPubkey) == 0 {
		return
	}
	// check that the client is authed to a pubkey in the owner follow list, this
	// relay is auth-to-read.
	r.Lock()
	defer r.Unlock()
	if len(r.Owners()) > 0 {
		for pk := range r.allFollowed {
			if bytes.Equal(authedPubkey, []byte(pk)) {
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

func (r *Relay) AcceptReq(c context.T, hr *http.Request, id []byte,
	ff *filters.T, authedPubkey []byte) (allowed *filters.T, ok bool, modified bool) {

	if r.PublicReadable { // && len(r.owners) == 0 {
		allowed = ff
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
				if fkk.IsDirectoryEvent() || (!fkk.IsPrivileged() && r.PublicReadable) {
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
			if len(f.Kinds.K) != len(allowedKinds.K) {
				modified = true
			}
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
	allowed = ff
	// client is permitted, pass through the filter so request/count processing does
	// not need logic and can just use the returned filter.
	// check that the client is authed to a pubkey in the owner follow list
	r.Lock()
	defer r.Unlock()
	if len(r.Owners()) > 0 {
		for pk := range r.allFollowed {
			if bytes.Equal(authedPubkey, []byte(pk)) {
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
// It also adds the allFollowed npubs of the follows.
func (r *Relay) CheckOwnerLists(c context.T) {
	if len(r.owners) > 0 {
		r.Lock()
		defer r.Unlock()
		var err error
		var evs []*event.T
		// need to search DB for moderator npub follow lists, allFollowed npubs are allowed access.
		if len(r.allFollowed) < 1 {
			// add the owners themselves of course
			for i := range r.owners {
				r.allFollowed[string(r.owners[i])] = struct{}{}
			}
			log.D.Ln("regenerating owners follow lists")
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.owners...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				r.ownersFollowLists = append(r.ownersFollowLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if bytes.Equal(t.Key(), []byte("p")) {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); chk.E(err) {
							continue
						}
						r.allFollowed[string(p)] = struct{}{}
						r.ownersFollowed[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
			// next, search for the follow lists of all on the follow list
			log.D.Ln("searching for owners follows follow lists")
			var followed []string
			for f := range r.allFollowed {
				followed = append(followed, f)
			}
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(followed...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				// we want to protect the follow lists of users as well so they also cannot be
				// deleted, only replaced.
				r.ownersFollowLists = append(r.ownersFollowLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if bytes.Equal(t.Key(), []byte("p")) {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); err != nil {
							continue
						}
						r.allFollowed[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		if len(r.muted) < 1 {
			log.D.Ln("regenerating owners mute lists")
			r.muted = make(map[string]struct{})
			if evs, err = r.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(r.owners...),
					Kinds: kinds.New(kind.MuteList)}); chk.E(err) {
			}
			for _, ev := range evs {
				r.ownersMuteLists = append(r.ownersMuteLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if bytes.Equal(t.Key(), []byte("p")) {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); chk.E(err) {
							continue
						}
						r.muted[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		// remove muted from the allFollowed list
		for m := range r.muted {
			for f := range r.allFollowed {
				if f == m {
					// delete muted element from allFollowed list
					delete(r.allFollowed, m)
				}
			}
		}
		log.I.F("%d allowed npubs, %d blocked", len(r.allFollowed), len(r.muted))
	}
}

func (r *Relay) AuthRequired() bool { return !r.PublicReadable || len(r.owners) > 0 }

// ServiceUrl returns the address of the relay to send back in auth responses.
// If auth is disabled this returns an empty string.
func (r *Relay) ServiceUrl(req *http.Request) (s string) {
	if !r.AuthRequired() {
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
