package realy

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/realy/listeners"
	"realy.lol/relay"
	"realy.lol/tag"
	"realy.lol/tags"
)

type Subscribe struct{ *Server }

func NewSubscribe(s *Server) (ep *Subscribe) { return &Subscribe{Server: s} }

type SubscribeInput struct {
	Auth string       `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"false" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
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

func (ep *Subscribe) RegisterSubscribe(api huma.API) {
	name := "Subscribe"
	description := "Subscribe for newly published events by author, kind or tags (empty also allowed)"
	path := "/subscribe"
	scopes := []string{"user", "read"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: generateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *SubscribeInput) (output *huma.StreamResponse, err error) {
		log.I.S(input)
		var f *filter.T
		if f, err = input.ToFilter(); chk.E(err) {
			err = huma.Error422UnprocessableEntity(err.Error())
			return
		}
		log.I.F("%s", f.Marshal(nil))
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		var valid bool
		var pubkey []byte
		valid, pubkey, err = httpauth.CheckAuth(r, ep.JWTVerifyFunc)
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
		// log.I.F("processing req\n%s\n", f.Serialize())
		allowed := filters.New(f)
		if accepter, ok := ep.relay.(relay.ReqAcceptor); ok {
			var accepted, modified bool
			allowed, accepted, modified = accepter.AcceptReq(ep.Ctx, r, nil, filters.New(f), pubkey)
			if !accepted {
				err = huma.Error401Unauthorized("auth to get access for this filter")
				return
			} else if modified {
				log.D.F("filter modified %s", allowed.F[0])
				// err = huma.Error401Unauthorized("returning results from modified filter; auth to get full access")
			}
		}
		if len(allowed.F) == 0 {
			err = huma.Error401Unauthorized("all kinds in event restricted; auth to get access for this filter")
			return
		}
		// log.I.F("allowed\n%s\n", allowed.Marshal(nil))
		if f.Kinds.IsPrivileged() {
			if auther, ok := ep.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
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
		// register the filter with the Listeners
		receiver := make(event.C, 32)
		s.Listeners.Hchan <- listeners.H{
			Ctx:      r.Context(),
			Receiver: receiver,
			Pubkey:   pubkey,
			Filter:   f,
		}
		output = &huma.StreamResponse{
			func(ctx huma.Context) {
				ctx.SetHeader("Content-Type", "text/event-stream")
				ctx.SetHeader("X-Accel-Buffering", "no")
				ctx.SetHeader("Cache-Control", "no-cache")
				w := ctx.BodyWriter()
				tick := time.NewTicker(time.Second)
			out:
				for {
					select {
					case <-tick.C:
						log.I.F("tick")
					case <-r.Context().Done():
						break out
					case ev := <-receiver:
						w.Write([]byte("data: "))
						w.Write(ev.Serialize())
						w.Write([]byte("\n\n"))
						if f, ok := ctx.BodyWriter().(http.Flusher); ok {
							f.Flush()
						} else {
							log.W.F("error: unable to flush")
						}
					}
				}
			},
		}

		return
	})
}
