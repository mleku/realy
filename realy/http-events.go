package realy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/cmd/realy/app"
	"realy.mleku.dev/context"
	"realy.mleku.dev/ec/schnorr"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/httpauth"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/sha256"
	"realy.mleku.dev/store"
	"realy.mleku.dev/tag"
)

// Events is a HTTP API method to retrieve a number of events from their event Ids.
type Events struct{ interfaces.Server }

// NewEvents creates a new Events with a provided Server.
func NewEvents(s interfaces.Server) (ep *Events) {
	return &Events{Server: s}
}

// EventsInput is the parameters for an Events HTTP API method. Basically an array of eventid.T.
type EventsInput struct {
	Auth string   `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Body []string `doc:"list of event Ids"`
}

// RegisterEvents is the implementation of the HTTP API for Events.
func (x *Events) RegisterEvents(api huma.API) {
	name := "Events"
	description := "Returns the full events from a list of event Ids as a line structured JSON. Auth required to fetch more than 1000 events, and if not enabled, 1000 is the limit."
	path := "/events"
	scopes := []string{"user", "read"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"events"},
		Description:   helpers.GenerateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *EventsInput) (output *huma.StreamResponse, err error) {
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
		if authrequired && len(pubkey) != schnorr.PubKeyBytesLen {
			err = huma.Error400BadRequest(
				"cannot process more than 1000 events in a request without being authenticated")
			return
		}
		if authrequired && valid {
			if len(x.Owners()) < 1 {
				err = huma.Error400BadRequest(
					"cannot process more than 1000 events in a request without auth enabled")
				return
			}
			if rl, ok := x.Relay().(*app.Relay); ok {
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
		if !valid {
			err = huma.Error401Unauthorized("Authorization header is invalid")
			return
		}
		sto := x.Storage()
		var evIds [][]byte
		for _, id := range input.Body {
			var idb []byte
			if idb, err = hex.Dec(id); chk.E(err) {
				err = huma.Error422UnprocessableEntity(err.Error())
				return
			}
			if len(idb) != sha256.Size {
				err = huma.Error422UnprocessableEntity(
					fmt.Sprintf("event Id must be 64 hex characters: '%s'", id))
			}
			evIds = append(evIds, idb)
		}
		if idsWriter, ok := sto.(store.GetIdsWriter); ok {
			output = &huma.StreamResponse{
				func(ctx huma.Context) {
					if err = idsWriter.FetchIds(x.Context(), tag.New(evIds...),
						ctx.BodyWriter()); chk.E(err) {
						return
					}
				},
			}
		}
		return
	})
}
