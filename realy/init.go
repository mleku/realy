package realy

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strings"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/kinds"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/config"
	"realy.mleku.dev/tag"
)

func getFirstTime() (s string) {
	b := make([]byte, 32)
	var err error
	if _, err = rand.Read(b); chk.E(err) {
		panic(err)
	}
	s = fmt.Sprintf("%x", b)
	return
}

func (s *Server) Configured() bool { return s.configured }

func (s *Server) Init() {
	var err error
	s.configurationMx.Lock()
	defer s.configurationMx.Unlock()
	if err = s.UpdateConfiguration(); chk.E(err) {
		return
	}
	if s.configuration == nil {
		s.configuration = &config.C{
			FirstTime:      getFirstTime(),
			BlockList:      nil,
			Owners:         nil,
			Admins:         nil,
			AuthRequired:   false,
			PublicReadable: true,
		}
		log.W.F(`first time configuration password: %s
    use with Authorization header to set at least 1 Admin`,
			s.configuration.FirstTime)
	} else {
		s.configured = true
	}
	// for _, src := range s.configuration.Owners {
	// 	if len(src) < 1 {
	// 		continue
	// 	}
	// 	dst := make([]byte, len(src)/2)
	// 	if _, err = hex.DecBytes(dst, []byte(src)); err != nil {
	// 		if dst, err = bech32encoding.NpubToBytes([]byte(src)); chk.E(err) {
	// 			continue
	// 		}
	// 	}
	// 	s.owners = append(s.owners, dst)
	// }
	if len(s.owners) > 0 {
		log.T.C(func() string {
			ownerIds := make([]string, len(s.owners))
			for i, npub := range s.owners {
				ownerIds[i] = hex.Enc(npub)
			}
			owners := strings.Join(ownerIds, ",")
			return fmt.Sprintf("owners %s", owners)
		})
		s.ZeroLists()
		s.CheckOwnerLists(context.Bg())
	}

}

func (s *Server) ZeroLists() {
	s.Lock()
	defer s.Unlock()
	s.followed = make(map[string]struct{})
	s.ownersFollowed = make(map[string]struct{})
	s.ownersFollowLists = s.ownersFollowLists[:0]
	s.muted = make(map[string]struct{})
	s.ownersMuteLists = s.ownersMuteLists[:0]
}

// CheckOwnerLists regenerates the owner follow and mute lists if they are empty.
//
// It also adds the followed npubs of the follows.
func (s *Server) CheckOwnerLists(c context.T) {
	s.Lock()
	defer s.Unlock()
	if len(s.owners) > 0 {
		var err error
		var evs []*event.T
		// need to search DB for moderator npub follow lists, followed npubs are allowed access.
		lf := len(s.followed)
		if lf < 1 {
			log.T.F("regenerating followed list")
			// add the owners themselves of course
			for i := range s.owners {
				log.I.F("added owner %0x to followed list", s.owners[i])
				s.followed[string(s.owners[i])] = struct{}{}
			}
			log.D.Ln("regenerating owners follow lists")
			if evs, err = s.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(s.owners...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				s.ownersFollowLists = append(s.ownersFollowLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if t.KeyString() == "p" {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); chk.E(err) {
							continue
						}
						s.followed[string(p)] = struct{}{}
						s.ownersFollowed[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
			// next, search for the follow lists of all on the follow list
			log.T.Ln("searching for owners follows follow lists")
			var followed []string
			for f := range s.followed {
				followed = append(followed, f)
			}
			if evs, err = s.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(followed...),
					Kinds: kinds.New(kind.FollowList)}); chk.E(err) {
			}
			for _, ev := range evs {
				// we want to protect the follow lists of users as well so they also cannot be
				// deleted, only replaced.
				s.ownersFollowLists = append(s.ownersFollowLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if bytes.Equal(t.Key(), []byte("p")) {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); err != nil {
							continue
						}
						s.followed[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		if len(s.muted) < 1 {
			log.D.Ln("regenerating owners mute lists")
			s.muted = make(map[string]struct{})
			if evs, err = s.Store.QueryEvents(c,
				&filter.T{Authors: tag.New(s.owners...),
					Kinds: kinds.New(kind.MuteList)}); chk.E(err) {
			}
			for _, ev := range evs {
				s.ownersMuteLists = append(s.ownersMuteLists, ev.Id)
				for _, t := range ev.Tags.ToSliceOfTags() {
					if bytes.Equal(t.Key(), []byte("p")) {
						var p []byte
						if p, err = hex.Dec(string(t.Value())); chk.E(err) {
							continue
						}
						s.muted[string(p)] = struct{}{}
					}
				}
			}
			evs = evs[:0]
		}
		// remove muted from the followed list
		for m := range s.muted {
			for f := range s.followed {
				if f == m {
					// delete muted element from followed list
					delete(s.followed, m)
				}
			}
		}
		log.I.F("%d allowed npubs, %d blocked", len(s.followed), len(s.muted))
	}
}
