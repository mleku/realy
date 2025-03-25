package realy

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

type SimpleFilter struct {
	Kinds   []int      `json:"kinds,omitempty" doc:"array of kind numbers to match on"`
	Authors []string   `json:"authors,omitempty" doc:"array of author pubkeys to match on (hex encoded)"`
	Tags    [][]string `json:"tags,omitempty" doc:"array of tags to match on (first key of each '#x' and terms to match from the second field of the event tag)"`
}

type Filter struct{ *Server }

func NewFilter(s *Server) (ep *Filter) { return &Filter{Server: s} }

type FilterInput struct {
	Auth  string       `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"false" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Since int64        `query:"since" doc:"timestamp of the oldest events to return (inclusive)"`
	Until int64        `query:"until" doc:"timestamp of the newest events to return (inclusive)"`
	Limit uint         `query:"limit" doc:"maximum number of results to return"`
	Sort  string       `query:"sort" enum:"asc,desc" default:"desc" doc:"sort order by created_at timestamp"`
	Body  SimpleFilter `body:"filter" doc:"filter criteria to match for events to return"`
}

func (fi FilterInput) ToFilter() (f *filter.T, err error) {
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
	if fi.Limit != 0 {
		f.Limit = &fi.Limit
	}
	if fi.Since != 0 {
		f.Since = timestamp.NewFromUnix(fi.Since)
	}
	if fi.Until != 0 {
		f.Until = timestamp.NewFromUnix(fi.Until)
	}
	return
}

type FilterOutput struct {
	Body []string `doc:"list of event Ids that mach the query in the sort order requested"`
}

func (ep *Filter) RegisterFilter(api huma.API) {
	name := "Filter"
	description := "Search for events and receive a sorted list of event Ids"
	path := "/filter"
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
	}, func(ctx context.T, input *FilterInput) (output *FilterOutput, err error) {
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
		var valid bool
		var pubkey []byte
		valid, pubkey, err = httpauth.CheckAuth(r, ep.JWTVerifyFunc)
		missing := !errors.Is(err, httpauth.ErrMissingKey)
		// if there is an error but not that the token is missing, or there is no error
		// but the signature is invalid, return error that request is unauthorized.
		if err != nil && !missing || err == nil && !valid {
			err = huma.Error401Unauthorized(
				fmt.Sprintf("invalid: %s", err.Error()))
			return
		}
		// log.I.F("processing req\n%s\n", f.Serialize())
		allowed := filters.New(f)
		if accepter, ok := ep.relay.(relay.ReqAcceptor); ok {
			var accepted, modified bool
			allowed, accepted, modified = accepter.AcceptReq(ep.Ctx, r, nil, filters.New(f), pubkey)
			if !accepted {
				err = huma.Error401Unauthorized("returning results from modified filter; auth to get full access")
				return
			} else if modified {
				log.D.F("filter modified %s", allowed.F[0])
				// err = huma.Error401Unauthorized("returning results from modified filter; auth to get full access")
			}
		}
		if allowed == nil {
			return
		}
		// log.I.F("allowed\n%s\n", allowed.Marshal(nil))
		if auther, ok := ep.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.IsPrivileged() {
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
		sto := ep.relay.Storage()
		var ok bool
		var quer store.Querier
		if quer, ok = sto.(store.Querier); !ok {
			err = huma.Error501NotImplemented("simple filter request not implemented")
			return
		}
		var evs []store.IdTsPk
		if evs, err = quer.QueryForIds(ep.Ctx, allowed.F[0]); chk.E(err) {
			err = huma.Error500InternalServerError("error querying for events", err)
			return
		}
		switch input.Sort {
		case "asc":
			sort.Slice(evs, func(i, j int) bool {
				return evs[i].Ts < evs[j].Ts
			})
		case "desc":
			sort.Slice(evs, func(i, j int) bool {
				return evs[i].Ts > evs[j].Ts
			})
		}
		if len(pubkey) > 0 {
			// remove events from results if we find the user's mute list, that are present
			// on this list
			var mutes event.Ts
			if mutes, err = sto.QueryEvents(ep.Ctx, &filter.T{Authors: tag.New(pubkey),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs [][]byte
				for _, ev := range mutes {
					for _, t := range ev.Tags.F() {
						if bytes.Equal(t.Key(), []byte("p")) {
							var p []byte
							if p, err = hex.Dec(string(t.Value())); chk.E(err) {
								continue
							}
							mutePubs = append(mutePubs, p)
						}
					}
				}
				var tmp []store.IdTsPk
			next:
				for _, ev := range evs {
					for _, pk := range mutePubs {
						if bytes.Equal(ev.Pub, pk) {
							continue next
						}
					}
					tmp = append(tmp, ev)
				}
				// log.I.F("done")
				evs = tmp
			}
		}
		output = &FilterOutput{}
		for _, ev := range evs {
			output.Body = append(output.Body, hex.Enc(ev.Id))
		}
		return
	})
}
