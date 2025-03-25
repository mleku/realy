package realy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/json"
	"realy.lol/tag"
)

type Events struct{ *Server }

func NewEvents(s *Server) (ep *Events) {
	return &Events{Server: s}
}

type EventsInput struct {
	Auth string   `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Body []string `doc:"list of event Ids"`
}

type EventsOutput struct{ RawBody []byte }

func (ep *Events) RegisterEvents(api huma.API) {
	name := "Events"
	description := "Returns the full events from a list of event Ids as a JSON array"
	path := "/events"
	scopes := []string{"user", "read"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"events"},
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *EventsInput) (output *EventsOutput, err error) {
		// log.I.S(input)
		r := ctx.Value("http-request").(*http.Request)
		w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		// s := ep.Server
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
		_ = pubkey
		sto := ep.relay.Storage()
		var evIds [][]byte
		for _, id := range input.Body {
			var pk []byte
			if pk, err = hex.Dec(id); chk.E(err) {
				err = huma.Error422UnprocessableEntity(err.Error())
				return
			}
			evIds = append(evIds, pk)
		}
		var evs event.Ts
		f := filter.T{IDs: tag.New(evIds...)}
		if evs, err = sto.QueryEvents(ep.Ctx, &f); chk.E(err) {
			err = huma.Error500InternalServerError(err.Error())
			return
		}
		if len(evs) == 0 {
			// no results, the end
			return
		}
		// log.I.S(evs)
		var res json.Array
		for _, ev := range evs {
			res.V = append(res.V, ev)
		}
		resB := res.Marshal(nil)
		_, err = w.Write(resB)
		// output = &EventsOutput{RawBody: resB}
		// log.I.F("%s", output.Body)
		return
	})
}
