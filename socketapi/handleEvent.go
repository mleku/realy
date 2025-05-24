package socketapi

import (
	"bytes"
	"strings"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/log"
	"realy.lol/realy/interfaces"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
)

func (a *A) HandleEvent(c context.T, req []byte, srv interfaces.Server,
	remote string) (msg []byte) {

	log.T.F("%s handleEvent %s", remote, req)
	var err error
	var ok bool
	var rem []byte
	sto := srv.Storage()
	if sto == nil {
		panic("no event store has been set to store event")
	}
	env := eventenvelope.NewSubmission()
	if rem, err = env.Unmarshal(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.T.F("%s extra '%s'", remote, rem)
	}
	log.I.F("authed pubkey: %0x", a.Listener.AuthedBytes())
	accept, notice, after := a.Server.AcceptEvent(c, env.T, a.Listener.Req(),
		a.Listener.AuthedBytes(), remote)
	log.T.F("%s accepted %v", remote, accept)
	if !accept {
		if err = a.HandleRejectEvent(env, notice); chk.E(err) {
			return
		}
		return
	}
	if err = a.VerifyEvent(env); chk.E(err) {
		return
	}
	if env.T.Kind.K == kind.Deletion.K {
		if err = a.CheckDelete(c, env, sto); chk.E(err) {
			return
		}
	}
	var reason []byte
	ok, reason = srv.AddEvent(c, env.T, a.Listener.Req(), a.Listener.AuthedBytes(), remote)
	log.T.F("%s <- event added %v", remote, ok)
	if err = okenvelope.NewFrom(env.Id(), ok, reason).Write(a.Listener); chk.E(err) {
		return
	}
	if after != nil {
		after()
	}
	return
}

func (a *A) VerifyEvent(env *eventenvelope.Submission) (err error) {
	if !bytes.Equal(env.GetIDBytes(), env.Id()) {
		if err = Ok.Invalid(a, env, "event id is computed incorrectly"); chk.E(err) {
			return
		}
		return
	}
	var ok bool
	if ok, err = env.Verify(); chk.T(err) {
		if err = Ok.Error(a, env, "failed to verify signature", err); chk.T(err) {
			return
		}
		return
	} else if !ok {
		if err = Ok.Error(a, env, "signature is invalid", err); chk.T(err) {
			return
		}
		return
	}
	return
}

func (a *A) HandleRejectEvent(env *eventenvelope.Submission, notice string) (err error) {
	if strings.Contains(notice, "mute") {
		if err = Ok.Blocked(a, env, notice); chk.E(err) {
			return
		}
	} else {
		if !a.Listener.AuthRequested() {
			a.Listener.RequestAuth()
			log.I.F("requesting auth from client %s", a.Listener.RealRemote())
		} else {
			log.I.F("requesting auth again from client %s", a.Listener.RealRemote())
		}
		if err = Ok.AuthRequired(a, env, "auth required for storing events"); chk.E(err) {
			return
		}
		if err = authenvelope.NewChallengeWith(a.Listener.Challenge()).Write(a.Listener); chk.T(err) {
			return
		}
		a.Listener.SetPendingEvent(env.T)
		return
	}
	if err = Ok.Invalid(a, env, notice); chk.E(err) {
		return
	}
	return
}

func (a *A) CheckDelete(c context.T, env *eventenvelope.Submission, sto store.I) (err error) {
	log.I.F("delete event\n%s", env.T.Serialize())
	for _, t := range env.Tags.ToSliceOfTags() {
		var res []*event.T
		if t.Len() >= 2 {
			switch {
			case bytes.Equal(t.Key(), []byte("e")):
				evId := make([]byte, sha256.Size)
				if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
					continue
				}
				res, err = sto.QueryEvents(c, &filter.T{IDs: tag.New(evId)})
				if err != nil {
					if err = Ok.Error(a, env, "failed to query for target event"); chk.T(err) {
						return
					}
					return
				}
				for i := range res {
					if res[i].Kind.Equal(kind.Deletion) {
						if err = Ok.Blocked(a, env,
							"not processing or storing delete event containing delete event references",
						); chk.E(err) {
							return
						}
						return
					}
					if !bytes.Equal(res[i].Pubkey, env.T.Pubkey) {
						if err = Ok.Blocked(a, env,
							"cannot delete other users' events (delete by e tag)",
						); chk.E(err) {
							return
						}
						return
					}
				}
			case bytes.Equal(t.Key(), []byte("a")):
				split := bytes.Split(t.Value(), []byte{':'})
				if len(split) != 3 {
					continue
				}
				var pk []byte
				if pk, err = hex.DecAppend(nil, split[1]); chk.E(err) {
					if err = Ok.Invalid(a, env,
						"delete event a tag pubkey value invalid: %s", t.Value()); chk.T(err) {
					}
					return
				}
				kin := ints.New(uint16(0))
				if _, err = kin.Unmarshal(split[0]); chk.E(err) {
					if err = Ok.Invalid(a, env,
						"delete event a tag kind value invalid: %s", t.Value()); chk.T(err) {
						return
					}
					return
				}
				kk := kind.New(kin.Uint16())
				if kk.Equal(kind.Deletion) {
					if err = Ok.Blocked(a, env, "delete event kind may not be deleted"); chk.E(err) {
						return
					}
					return
				}
				if !kk.IsParameterizedReplaceable() {
					if err = Ok.Error(a, env,
						"delete tags with a tags containing non-parameterized-replaceable events cannot be processed"); chk.E(err) {
						return
					}
					return
				}
				if !bytes.Equal(pk, env.T.Pubkey) {
					log.I.S(pk, env.T.Pubkey, env.T)
					if err = Ok.Blocked(a, env,
						"cannot delete other users' events (delete by a tag)"); chk.E(err) {
						return
					}
					return
				}
				f := filter.New()
				f.Kinds.K = []*kind.T{kk}
				f.Authors.Append(pk)
				f.Tags.AppendTags(tag.New([]byte{'#', 'd'}, split[2]))
				if res, err = sto.QueryEvents(c, f); err != nil {
					if err = Ok.Error(a, env,
						"failed to query for target event"); chk.T(err) {
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
			var skip bool
			if skip, err = a.ProcessDelete(c, target, env, sto); skip {
				continue
			} else if err != nil {
				return
			}
		}
		res = nil
	}
	if err = okenvelope.NewFrom(env.Id(), true).Write(a.Listener); chk.E(err) {
		return
	}
	return
}

func (a *A) ProcessDelete(c context.T, target *event.T, env *eventenvelope.Submission,
	sto store.I) (skip bool, err error) {
	if target.Kind.K == kind.Deletion.K {
		if err = Ok.Error(a, env, "cannot delete delete event %s", env.Id); chk.E(err) {
			return
		}
	}
	if target.CreatedAt.Int() > env.T.CreatedAt.Int() {
		if err = Ok.Error(a, env,
			"not deleting\n%d%\nbecause delete event is older\n%d",
			target.CreatedAt.Int(), env.T.CreatedAt.Int()); chk.E(err) {
			return
		}
		skip = true
	}
	if !bytes.Equal(target.Pubkey, env.Pubkey) {
		if err = Ok.Error(a, env, "only author can delete event"); chk.E(err) {
			return
		}
		return
	}
	if err = sto.DeleteEvent(c, target.EventId()); chk.T(err) {
		if err = Ok.Error(a, env, err.Error()); chk.T(err) {
			return
		}
		return
	}
	return
}
