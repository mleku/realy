package realy

import (
	"bytes"
	"strings"

	"realy.lol/context"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

func (s *Server) handleEvent(c context.T, ws *web.Socket, req []byte, sto store.I) (msg []byte) {
	log.T.F("handleEvent %s %s", ws.RealRemote(), req)
	var err error
	var ok bool
	var rem []byte
	advancedDeleter, _ := sto.(relay.AdvancedDeleter)
	env := eventenvelope.NewSubmission()
	if rem, err = env.Unmarshal(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	accept, notice, after := s.relay.AcceptEvent(c, env.T, ws.Req(),
		ws.RealRemote(), ws.AuthedBytes())
	if !accept {
		if strings.Contains(notice, "mute") {
			if err = okenvelope.NewFrom(env.ID, false,
				normalize.Blocked.F(notice)).Write(ws); chk.T(err) {
			}
		} else {
			log.I.F("AUTHING")
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
				if !ws.AuthRequested() {
					ws.RequestAuth()
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.T(err) {
					}
					log.I.F("requesting auth from client %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.T(err) {
						return
					}
					return
				} else {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.AuthRequired.F("auth required for storing events")).Write(ws); chk.T(err) {
					}
					log.I.F("requesting auth again from client %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.T(err) {
						return
					}
					return
				}
			} else {
				log.W.F("didn't find authentication method")
			}
		}
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F(notice)).Write(ws); chk.T(err) {
		}
		return
	}
	if !bytes.Equal(env.GetIDBytes(), env.ID) {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F("event id is computed incorrectly")).Write(ws); chk.E(err) {
			return
		}
		return
	}
	if ok, err = env.Verify(); chk.T(err) {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("failed to verify signature")).Write(ws); chk.E(err) {
			return
		}
	} else if !ok {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("signature is invalid")).Write(ws); chk.E(err) {
			return
		}
		return
	}
	storage := s.relay.Storage()
	if storage == nil {
		panic("no event store has been set to store event")
	}
	if env.T.Kind.K == kind.Deletion.K {
		log.I.F("delete event\n%s", env.T.Serialize())
		for _, t := range env.Tags.Value() {
			var res []*event.T
			if t.Len() >= 2 {
				switch {
				case bytes.Equal(t.Key(), []byte("e")):
					evId := make([]byte, sha256.Size)
					if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
						continue
					}
					res, err = storage.QueryEvents(c, &filter.T{IDs: tag.New(evId)})
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
					for i := range res {
						if res[i].Kind.Equal(kind.Deletion) {
							if err = okenvelope.NewFrom(env.ID, false,
								normalize.Blocked.F("not processing or storing delete event containing delete event references")).Write(ws); chk.E(err) {
								return
							}
						}
						if !bytes.Equal(res[i].PubKey, env.T.PubKey) {
							if err = okenvelope.NewFrom(env.ID, false,
								normalize.Blocked.F("cannot delete other users' events (delete by e tag)")).Write(ws); chk.E(err) {
								return
							}
						}
					}
				case bytes.Equal(t.Key(), []byte("a")):
					split := bytes.Split(t.Value(), []byte{':'})
					if len(split) != 3 {
						continue
					}
					var pk []byte
					if pk, err = hex.DecAppend(nil, split[1]); chk.E(err) {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Invalid.F("delete event a tag pubkey value invalid: %s",
								t.Value())).Write(ws); chk.E(err) {
							return
						}
						return
					}
					kin := ints.New(uint16(0))
					if _, err = kin.Unmarshal(split[0]); chk.E(err) {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Invalid.F("delete event a tag kind value invalid: %s",
								t.Value())).Write(ws); chk.E(err) {
							return
						}
						return
					}
					kk := kind.New(kin.Uint16())
					if kk.Equal(kind.Deletion) {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Blocked.F("delete event kind may not be deleted")).Write(ws); chk.E(err) {
							return
						}
					}
					if !kk.IsParameterizedReplaceable() {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("delete tags with a tags containing non-parameterized-replaceable events cannot be processed")).Write(ws); chk.E(err) {
							return
						}
					}
					if !bytes.Equal(pk, env.T.PubKey) {
						log.I.S(pk, env.T.PubKey, env.T)
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Blocked.F("cannot delete other users' events (delete by a tag)")).Write(ws); chk.E(err) {
							return
						}
					}
					f := filter.New()
					f.Kinds.K = []*kind.T{kk}
					// aut := make(by, 0, len(pk)/2)
					// if aut, err = hex.DecAppend(aut, pk); chk.E(err) {
					// 	return
					// }
					f.Authors.Append(pk)
					f.Tags.AppendTags(tag.New([]byte{'#', 'd'}, split[2]))
					res, err = storage.QueryEvents(c, f)
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
				}
			}
			if len(res) < 1 {
				continue
			}
			var resTmp []*event.T
			for _, v := range res {
				if env.T.CreatedAt.U64() >= v.CreatedAt.U64() {
					resTmp = append(resTmp, v)
				}
			}
			res = resTmp
			for _, target := range res {
				if target.Kind.K == kind.Deletion.K {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F("cannot delete delete event %s",
							env.ID)).Write(ws); chk.E(err) {
						return
					}
				}
				if target.CreatedAt.Int() > env.T.CreatedAt.Int() {
					log.I.F("not deleting\n%d%\nbecause delete event is older\n%d",
						target.CreatedAt.Int(), env.T.CreatedAt.Int())
					continue
				}
				if !bytes.Equal(target.PubKey, env.PubKey) {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F("only author can delete event")).Write(ws); chk.E(err) {
						return
					}
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(c, t.Value(), env.PubKey)
				}
				if err = sto.DeleteEvent(c, target.EventID()); chk.T(err) {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
						return
					}
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.Value(), env.PubKey)
				}
			}
			res = nil
		}
		if err = okenvelope.NewFrom(env.ID, true).Write(ws); chk.E(err) {
			return
		}
	}
	var reason []byte
	ok, reason = s.addEvent(c, s.relay, env.T, ws.Req(), ws.RealRemote(), ws.AuthedBytes())
	log.I.F("event added %v, %s", ok, reason)
	if err = okenvelope.NewFrom(env.ID, ok, reason).Write(ws); chk.E(err) {
		return
	}
	if after != nil {
		after()
	}
	return
}
