package openapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/httpauth"
	"realy.mleku.dev/ints"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/sha256"
	"realy.mleku.dev/tag"
)

// EventInput is the parameters for the Event HTTP API method.
type EventInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false"`
	RawBody []byte
}

// EventOutput is the return parameters for the HTTP API Event method.
type EventOutput struct{ Body string }

// RegisterEvent is the implementatino of the HTTP API Event method.
func (x *Operations) RegisterEvent(api huma.API) {
	name := "Event"
	description := "Submit an event"
	path := x.path + "/event"
	scopes := []string{"user", "write"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *EventInput) (output *EventOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		remote := helpers.GetRemoteFromReq(r)
		ev := &event.T{}
		if _, err = ev.Unmarshal(input.RawBody); chk.E(err) {
			err = huma.Error406NotAcceptable(err.Error())
			return
		}
		var ok bool
		sto := x.Storage()
		if sto == nil {
			panic("no event store has been set to store event")
		}
		var pubkey []byte
		if x.Server.AuthRequired() || !x.Server.PublicReadable() {
			var valid bool
			valid, pubkey, err = httpauth.CheckAuth(r)
			// missing := !errors.Is(err, httpauth.ErrMissingKey)
			// if there is an error but not that the token is missing, or there is no error
			// but the signature is invalid, return error that request is unauthorized.
			if err != nil && !errors.Is(err, httpauth.ErrMissingKey) {
				err = huma.Error400BadRequest(err.Error())
				return
			}
			err = nil
			if !valid {
				err = huma.Error401Unauthorized("Authorization header is invalid")
				return
			}
		}
		// if there was auth, or no auth, check the relay policy allows accepting the
		// event (no auth with auth required or auth not valid for action can apply
		// here).
		accept, notice, after := x.AcceptEvent(ctx, ev, r, pubkey, remote)
		if !accept {
			err = huma.Error401Unauthorized(notice)
			return
		}
		if !bytes.Equal(ev.GetIDBytes(), ev.Id) {
			err = huma.Error400BadRequest("event id is computed incorrectly")
			return
		}
		if ok, err = ev.Verify(); chk.T(err) {
			err = huma.Error400BadRequest("failed to verify signature")
			return
		} else if !ok {
			err = huma.Error400BadRequest("signature is invalid")
			return
		}
		if ev.Kind.K == kind.Deletion.K {
			log.I.F("delete event\n%s", ev.Serialize())
			for _, t := range ev.Tags.ToSliceOfTags() {
				var res []*event.T
				if t.Len() >= 2 {
					switch {
					case bytes.Equal(t.Key(), []byte("e")):
						evId := make([]byte, sha256.Size)
						if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
							continue
						}
						res, err = sto.QueryEvents(ctx, &filter.T{IDs: tag.New(evId)})
						if err != nil {
							err = huma.Error500InternalServerError(err.Error())
							return
						}
						for i := range res {
							if res[i].Kind.Equal(kind.Deletion) {
								err = huma.Error409Conflict("not processing or storing delete event containing delete event references")
							}
							if !bytes.Equal(res[i].Pubkey, ev.Pubkey) {
								err = huma.Error409Conflict("cannot delete other users' events (delete by e tag)")
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
							err = huma.Error400BadRequest(fmt.Sprintf("delete event a tag pubkey value invalid: %s",
								t.Value()))
							return
						}
						kin := ints.New(uint16(0))
						if _, err = kin.Unmarshal(split[0]); chk.E(err) {
							err = huma.Error400BadRequest(fmt.Sprintf("delete event a tag kind value invalid: %s",
								t.Value()))
							return
						}
						kk := kind.New(kin.Uint16())
						if kk.Equal(kind.Deletion) {
							err = huma.Error403Forbidden("delete event kind may not be deleted")
							return
						}
						if !kk.IsParameterizedReplaceable() {
							err = huma.Error403Forbidden("delete tags with a tags containing non-parameterized-replaceable events cannot be processed")
							return
						}
						if !bytes.Equal(pk, ev.Pubkey) {
							log.I.S(pk, ev.Pubkey, ev)
							err = huma.Error403Forbidden("cannot delete other users' events (delete by a tag)")
							return
						}
						f := filter.New()
						f.Kinds.K = []*kind.T{kk}
						f.Authors.Append(pk)
						f.Tags.AppendTags(tag.New([]byte{'#', 'd'}, split[2]))
						res, err = sto.QueryEvents(ctx, f)
						if err != nil {
							err = huma.Error500InternalServerError(err.Error())
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
						err = huma.Error403Forbidden(fmt.Sprintf(
							"cannot delete delete event %s", ev.Id))
						return
					}
					if target.CreatedAt.Int() > ev.CreatedAt.Int() {
						// todo: shouldn't this be an error?
						log.I.F("not deleting\n%d%\nbecause delete event is older\n%d",
							target.CreatedAt.Int(), ev.CreatedAt.Int())
						continue
					}
					if !bytes.Equal(target.Pubkey, ev.Pubkey) {
						err = huma.Error403Forbidden("only author can delete event")
						return
					}
					if err = sto.DeleteEvent(ctx, target.EventId()); chk.T(err) {
						err = huma.Error500InternalServerError(err.Error())
						return
					}
				}
				res = nil
			}
			return
		}
		var reason []byte
		ok, reason = x.AddEvent(ctx, ev, r, pubkey, remote)
		// return the response whether true or false and any reason if false
		if ok {
		} else {
			err = huma.Error500InternalServerError(string(reason))
		}
		if after != nil {
			// do this in the background and let the http response close
			go after()
		}
		output = &EventOutput{"event accepted"}
		return
	})
}
