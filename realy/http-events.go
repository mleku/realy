package realy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/ec/schnorr"
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
	Auth string   `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"false" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Body []string `doc:"list of event Ids"`
}

type EventsOutput struct {
	RawBody []byte `doc:"the requested events as an array of events in JSON wire format"`
}

func (ep *Events) RegisterEvents(api huma.API) {
	name := "Events"
	description := "Returns the full events from a list of event Ids as a JSON array. Auth required to fetch more than 1000 events, and if not enabled, is not available"
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
		if len(input.Body) == 10000 {
			err = huma.Error400BadRequest(
				"cannot process more than 10000 events in a request")
			return

		}
		var authrequired bool
		if len(input.Body) > 1000 {
			authrequired = true
		}
		r := ctx.Value("http-request").(*http.Request)
		w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		s := ep.Server
		var valid bool
		var pubkey []byte
		valid, pubkey, err = httpauth.CheckAuth(r, ep.JWTVerifyFunc)
		if authrequired && len(pubkey) != schnorr.PubKeyBytesLen {
			err = huma.Error400BadRequest(
				"cannot process more than 1000 events in a request without being authenticated")
			return
		}
		if authrequired && valid {
			if len(s.owners) < 1 {
				err = huma.Error400BadRequest(
					"cannot process more than 1000 events in a request without auth enabled")
				return
			}
			if rl, ok := s.relay.(*app.Relay); ok {
				rl.Lock()
				// we only allow the first level of the allowed users this kind of access
				if _, ok = rl.OwnersFollowed[string(pubkey)]; !ok {
					err = huma.Error403Forbidden(
						fmt.Sprintf(
							"authenticated user %0x does not have permission for this request (owners can use export)",
							pubkey))
					return
				}
			}
		}
		// if there is an error but not that the token is missing, or there is no error
		// but the signature is invalid, return error that request is unauthorized.
		if err != nil && !errors.Is(err, httpauth.ErrMissingKey) {
			err = huma.Error400BadRequest(err.Error())
			return
		}
		if !valid {
			err = huma.Error401Unauthorized("Authorization header is invalid")
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
