package openapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/httpauth"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/kinds"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/tags"
)

type SubscribeInput struct {
	Auth   string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false"`
	Accept string `header:"Accept" default:"text/event-stream" enum:"text/event-stream" required:"true"`
	// ContentType string       `header:"Content-Type" default:"text/event-stream" enum:"text/event-stream" required:"true"`
	Body SimpleFilter `body:"filter" doc:"filter criteria to match for events to return"`
}

func (fi SubscribeInput) ToFilter() (f *filter.T, err error) {
	f = filter.New()
	var ks []*kind.T
	for _, k := range fi.Body.Kinds {
		ks = append(ks, kind.New(k))
	}
	f.Kinds = kinds.New(ks...)
	var as [][]byte
	for _, a := range fi.Body.Authors {
		var b []byte
		if b, err = hex.Dec(a); chk.E(err) {
			return
		}
		as = append(as, b)
	}
	f.Authors = tag.New(as...)
	var ts []*tag.T
	for _, t := range fi.Body.Tags {
		ts = append(ts, tag.New(t...))
	}
	f.Tags = tags.New(ts...)
	return
}

func (x *Operations) RegisterSubscribe(api huma.API) {
	name := "Subscribe"
	description := "Subscribe for newly published events by author, kind or tags; empty also allowed, which just sends all incoming events - uses Server Sent Events format for compatibility with standard libraries."
	path := "/subscribe"
	scopes := []string{"user", "read"}
	method := http.MethodPost
	sse.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	},
		map[string]any{
			"event": event.J{},
		},
		func(ctx context.T, input *SubscribeInput, send sse.Sender) {
			log.I.S(input)
			var err error
			var f *filter.T
			if f, err = input.ToFilter(); chk.E(err) {
				err = huma.Error422UnprocessableEntity(err.Error())
				return
			}
			log.I.F("%s", f.Marshal(nil))
			r := ctx.Value("http-request").(*http.Request)
			rr := helpers.GetRemoteFromReq(r)
			var valid bool
			var pubkey []byte
			valid, pubkey, err = httpauth.CheckAuth(r)
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
			allowed := filters.New(f)
			var accepted, modified bool
			allowed, accepted, modified = x.Relay().AcceptReq(x.Context(), r, nil,
				filters.New(f),
				pubkey)
			if !accepted {
				err = huma.Error401Unauthorized("auth to get access for this filter")
				return
			} else if modified {
				log.D.F("filter modified %s", allowed.F[0])
			}
			if len(allowed.F) == 0 {
				err = huma.Error401Unauthorized("all kinds in event restricted; auth to get access for this filter")
				return
			}
			if f.Kinds.IsPrivileged() {
				if x.Relay().AuthRequired() {
					log.T.F("privileged request\n%s", f.Serialize())
					senders := f.Authors
					receivers := f.Tags.GetAll(tag.New("#p"))
					switch {
					case len(pubkey) == 0:
						err = huma.Error401Unauthorized("auth required for processing request due to presence of privileged kinds (DMs, app specific data)")
						return
					case senders.Contains(pubkey) || receivers.ContainsAny([]byte("#p"),
						tag.New(pubkey)):
						log.T.F("user %0x from %s allowed to query for privileged event",
							pubkey, rr)
					default:
						err = huma.Error403Forbidden(fmt.Sprintf(
							"authenticated user %0x does not have authorization for "+
								"requested filters", pubkey))
					}
				}
			}
			// register the filter with the listeners
			receiver := make(event.C, 32)
			x.Publisher().Receive(&H{
				Ctx:      r.Context(),
				Receiver: receiver,
				Pubkey:   pubkey,
				Filter:   f,
			})
		out:
			for {
				select {
				case <-r.Context().Done():
					break out
				case ev := <-receiver:
					if err = send.Data(ev.ToEventJ()); chk.E(err) {
					}
				}
			}
			return
		})
}
