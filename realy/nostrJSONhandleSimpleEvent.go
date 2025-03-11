package realy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/relay"
	"realy.lol/sha256"
	"realy.lol/tag"
)

const (
	NA  = http.StatusNotAcceptable
	NI  = http.StatusNotImplemented
	ERR = http.StatusInternalServerError
)

func (s *Server) handleSimpleEvent(h Handler) {
	log.I.F("event")
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
	if valid, pubkey, err = httpauth.CheckAuth(h.Request, s.JWTVerifyFunc); chk.E(err) {
		return
	}
	rw := h.ResponseWriter
	if !valid {
		http.Error(rw,
			fmt.Sprintf("invalid: %s", err.Error()), NA)
	}
	rr := GetRemoteFromReq(h.Request)
	c := context.Bg()
	accept, notice, after := s.relay.AcceptEvent(c, ev, h.Request, rr, pubkey)
	if !accept {
		http.Error(rw, notice, NA)
		return
	}
	if !bytes.Equal(ev.GetIDBytes(), ev.ID) {
		http.Error(rw,
			"Event id is computed incorrectly", NA)
		return
	}
	if ok, err = ev.Verify(); chk.T(err) {
		http.Error(rw,
			"failed to verify signature", NA)
		return
	} else if !ok {
		http.Error(rw,
			"signature is invalid", NA)
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
						http.Error(rw,
							err.Error(), ERR)
						return
					}
					for i := range res {
						if res[i].Kind.Equal(kind.Deletion) {
							http.Error(rw,
								"not processing or storing delete event containing delete event references",
								NA)
						}
						if !bytes.Equal(res[i].PubKey, ev.PubKey) {
							http.Error(rw,
								"cannot delete other users' events (delete by e tag)",
								NA)
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
						http.Error(rw,
							fmt.Sprintf("delete event a tag pubkey value invalid: %s",
								t.Value()), NA)
						return
					}
					kin := ints.New(uint16(0))
					if _, err = kin.Unmarshal(split[0]); chk.E(err) {
						http.Error(rw,
							fmt.Sprintf("delete event a tag kind value invalid: %s",
								t.Value()), NA)
						return
					}
					kk := kind.New(kin.Uint16())
					if kk.Equal(kind.Deletion) {
						http.Error(rw,
							"delete event kind may not be deleted",
							NA)
						return
					}
					if !kk.IsParameterizedReplaceable() {
						http.Error(rw,
							"delete tags with a tags containing non-parameterized-replaceable events cannot be processed",
							NA)
						return
					}
					if !bytes.Equal(pk, ev.PubKey) {
						log.I.S(pk, ev.PubKey, ev)
						http.Error(rw,
							"cannot delete other users' events (delete by a tag)",
							NA)
						return
					}
					f := filter.New()
					f.Kinds.K = []*kind.T{kk}
					f.Authors.Append(pk)
					f.Tags.AppendTags(tag.New([]byte{'#', 'd'}, split[2]))
					res, err = storage.QueryEvents(c, f)
					if err != nil {
						http.Error(rw, err.Error(), ERR)
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
					http.Error(rw,
						fmt.Sprintf("cannot delete delete event %s",
							ev.ID), NA)
					return
				}
				if target.CreatedAt.Int() > ev.CreatedAt.Int() {
					// todo: shouldn't this be an error?
					log.I.F("not deleting\n%d%\nbecause delete event is older\n%d",
						target.CreatedAt.Int(), ev.CreatedAt.Int())
					continue
				}
				if !bytes.Equal(target.PubKey, ev.PubKey) {
					http.Error(rw,
						"only author can delete event",
						NA)
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(c, t.Value(), ev.PubKey)
				}
				if err = sto.DeleteEvent(c, target.EventID()); chk.T(err) {
					http.Error(rw,
						err.Error(), ERR)
					return
				}
				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.Value(), ev.PubKey)
				}
			}
			res = nil
		}
		http.Error(rw, "", http.StatusOK)
		return
	}
	var reason []byte
	ok, reason = s.addEvent(c, s.relay, ev, h.Request, rr, pubkey)
	// return the response whether true or false and any reason if false
	if ok {
		http.Error(rw, "", http.StatusOK)
	} else {
		http.Error(rw, string(reason), ERR)
	}
	if after != nil {
		// do this in the background and let the http response close
		go after()
	}
	return
}
