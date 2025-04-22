package openapi

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/danielgtaylor/huma/v2"

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
	"realy.mleku.dev/store"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/tags"
	"realy.mleku.dev/timestamp"
)

// SimpleFilter is the main parts of a filter.T that relate to event store indexes.
type SimpleFilter struct {
	Kinds   []int      `json:"kinds,omitempty" doc:"array of kind numbers to match on"`
	Authors []string   `json:"authors,omitempty" doc:"array of author pubkeys to match on (hex encoded)"`
	Tags    [][]string `json:"tags,omitempty" doc:"array of tags to match on (first key of each '#x' and terms to match from the second field of the event tag)"`
}

// FilterInput is the parameters for a Filter HTTP API call.
type FilterInput struct {
	Auth  string       `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false"`
	Since int64        `query:"since" doc:"timestamp of the oldest events to return (inclusive)"`
	Until int64        `query:"until" doc:"timestamp of the newest events to return (inclusive)"`
	Limit uint         `query:"limit" doc:"maximum number of results to return"`
	Sort  string       `query:"sort" enum:"asc,desc" default:"desc" doc:"sort order by created_at timestamp"`
	Body  SimpleFilter `body:"filter" doc:"filter criteria to match for events to return"`
}

// ToFilter converts a SimpleFilter input to a regular nostr filter.T.
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
		f.Since = timestamp.New(fi.Since)
	}
	if fi.Until != 0 {
		f.Until = timestamp.New(fi.Until)
	}
	return
}

// FilterOutput is a list of event Ids that match the query in the sort order requested.
type FilterOutput struct {
	Body []string `doc:"list of event Ids that mach the query in the sort order requested"`
}

// RegisterFilter is the implementation of the HTTP API Filter method.
func (x *Operations) RegisterFilter(api huma.API) {
	name := "Filter"
	description := "Search for events and receive a sorted list of event Ids (one of authors, kinds or tags must be present)"
	path := "/filter"
	scopes := []string{"user", "read"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: helpers.GenerateDescription(description, scopes),
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
		rr := helpers.GetRemoteFromReq(r)
		if len(input.Body.Authors) < 1 && len(input.Body.Kinds) < 1 && len(input.Body.Tags) < 1 {
			err = huma.Error400BadRequest(
				"cannot process filter with none of Authors/Kinds/Tags")
			return
		}
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
			filters.New(f), pubkey)
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
		sto := x.Storage()
		var ok bool
		var quer store.Querier
		if quer, ok = sto.(store.Querier); !ok {
			err = huma.Error501NotImplemented("simple filter request not implemented")
			return
		}
		var evs []store.IdTsPk
		if evs, err = quer.QueryForIds(x.Context(), allowed.F[0]); chk.E(err) {
			err = huma.Error500InternalServerError("error querying for events", err)
			return
		}
		if input.Limit > 0 {
			evs = evs[:input.Limit]
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
			if mutes, err = sto.QueryEvents(x.Context(), &filter.T{Authors: tag.New(pubkey),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs [][]byte
				for _, ev := range mutes {
					for _, t := range ev.Tags.ToSliceOfTags() {
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
				// log.I.ToSliceOfBytes("done")
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
