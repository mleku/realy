package realy

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"realy.lol/context"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/sha256"
	"realy.lol/tag"
)

func GetRemoteFromReq(r *http.Request) (rr string) {
	// reverse proxy should populate this field so we see the remote not the proxy
	rr = r.Header.Get("X-Forwarded-For")
	if rr != "" {
		splitted := strings.Split(rr, " ")
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
		// in case upstream doesn't set this or we are directly listening instead of
		// via reverse proxy or just if the header field is missing, put the
		// connection remote address into the websocket state data.
		if rr == "" {
			rr = r.RemoteAddr
		}
	}
	return
}

func (s *Server) handleSimpleEvent(h Handler) {
	var err error
	var ok bool
	sto := s.relay.Storage()
	var req []byte
	if req, err = io.ReadAll(h.Request.Body); chk.E(err) {
		return
	}
	advancedDeleter, _ := sto.(relay.AdvancedDeleter)
	ev := event.New()
	if req, err = ev.Unmarshal(req); chk.T(err) {
		return
	}
	var valid bool
	var pubkey []byte
	if valid, pubkey, err = httpauth.ValidateRequest(h.Request, s.JWTVerifyFunc); chk.E(err) {
		return
	}
	if !valid {
		return
	}
	rr := GetRemoteFromReq(h.Request)
	c := context.Bg()
	rw := h.ResponseWriter
	accept, notice, after := s.relay.AcceptEvent(c, ev, h.Request, rr, pubkey)
	if !accept {
		if strings.Contains(notice, "mute") {
			if err = okenvelope.NewFrom(ev.ID, false,
				normalize.Blocked.F(notice)).Write(rw); chk.T(err) {
			}
			return
		}
		if err = okenvelope.NewFrom(ev.ID, false,
			normalize.Invalid.F(notice)).Write(rw); chk.T(err) {
		}
		return
	}
	if !bytes.Equal(ev.GetIDBytes(), ev.ID) {
		if err = okenvelope.NewFrom(ev.ID, false,
			normalize.Invalid.F("event id is computed incorrectly")).Write(rw); chk.E(err) {
			return
		}
		return
	}
	if ok, err = ev.Verify(); chk.T(err) {
		if err = okenvelope.NewFrom(ev.ID, false,
			normalize.Error.F("failed to verify signature")).Write(rw); chk.E(err) {
			return
		}
	} else if !ok {
		if err = okenvelope.NewFrom(ev.ID, false,
			normalize.Error.F("signature is invalid")).Write(rw); chk.E(err) {
			return
		}
		return
	}
	storage := s.relay.Storage()
	if storage == nil {
		panic("no event store has been set to store event")
	}
	if ev.Kind.K == kind.Deletion.K {
		log.I.F("delete event\n%s", ev.Serialize())
		for _, t := range ev.Tags.Value() {
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
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Error.F("failed to query for target event")).Write(rw); chk.E(err) {
							return
						}
						return
					}
					for i := range res {
						if res[i].Kind.Equal(kind.Deletion) {
							if err = okenvelope.NewFrom(ev.ID, false,
								normalize.Blocked.F("not processing or storing delete event containing delete event references")).Write(rw); chk.E(err) {
								return
							}
						}
						if !bytes.Equal(res[i].PubKey, ev.PubKey) {
							if err = okenvelope.NewFrom(ev.ID, false,
								normalize.Blocked.F("cannot delete other users' events (delete by e tag)")).Write(rw); chk.E(err) {
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
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Invalid.F("delete event a tag pubkey value invalid: %s",
								t.Value())).Write(rw); chk.E(err) {
							return
						}
						return
					}
					kin := ints.New(uint16(0))
					if _, err = kin.Unmarshal(split[0]); chk.E(err) {
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Invalid.F("delete event a tag kind value invalid: %s",
								t.Value())).Write(rw); chk.E(err) {
							return
						}
						return
					}
					kk := kind.New(kin.Uint16())
					if kk.Equal(kind.Deletion) {
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Blocked.F("delete event kind may not be deleted")).Write(rw); chk.E(err) {
							return
						}
					}
					if !kk.IsParameterizedReplaceable() {
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Error.F("delete tags with a tags containing non-parameterized-replaceable events cannot be processed")).Write(rw); chk.E(err) {
							return
						}
					}
					if !bytes.Equal(pk, ev.PubKey) {
						log.I.S(pk, ev.PubKey, ev)
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Blocked.F("cannot delete other users' events (delete by a tag)")).Write(rw); chk.E(err) {
							return
						}
					}
					f := filter.New()
					f.Kinds.K = []*kind.T{kk}
					f.Authors.Append(pk)
					f.Tags.AppendTags(tag.New([]byte{'#', 'd'}, split[2]))
					res, err = storage.QueryEvents(c, f)
					if err != nil {
						if err = okenvelope.NewFrom(ev.ID, false,
							normalize.Error.F("failed to query for target event")).Write(rw); chk.E(err) {
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
				if ev.CreatedAt.U64() >= v.CreatedAt.U64() {
					resTmp = append(resTmp, v)
				}
			}
			res = resTmp
			for _, target := range res {
				if target.Kind.K == kind.Deletion.K {
					if err = okenvelope.NewFrom(ev.ID, false,
						normalize.Error.F("cannot delete delete event %s",
							ev.ID)).Write(rw); chk.E(err) {
						return
					}
				}
				if target.CreatedAt.Int() > ev.CreatedAt.Int() {
					log.I.F("not deleting\n%d%\nbecause delete event is older\n%d",
						target.CreatedAt.Int(), ev.CreatedAt.Int())
					continue
				}
				if !bytes.Equal(target.PubKey, ev.PubKey) {
					if err = okenvelope.NewFrom(ev.ID, false,
						normalize.Error.F("only author can delete event")).Write(rw); chk.E(err) {
						return
					}
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(c, t.Value(), ev.PubKey)
				}
				if err = sto.DeleteEvent(c, target.EventID()); chk.T(err) {
					if err = okenvelope.NewFrom(ev.ID, false,
						normalize.Error.F(err.Error())).Write(rw); chk.E(err) {
						return
					}
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.Value(), ev.PubKey)
				}
			}
			res = nil
		}
		if err = okenvelope.NewFrom(ev.ID, true).Write(rw); chk.E(err) {
			return
		}
	}
	var reason []byte
	ok, reason = s.addEvent(c, s.relay, ev, h.Request, rr, pubkey)
	if err = okenvelope.NewFrom(ev.ID, ok, reason).Write(rw); chk.E(err) {
		return
	}
	if after != nil {
		after()
	}
	return
}
