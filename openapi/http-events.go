package openapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rickb777/acceptable/header"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/httpauth"
	"realy.lol/realy/helpers"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
)

// EventsInput is the parameters for an Events HTTP API method. Basically an array of eventid.T.
type EventsInput struct {
	Auth   string   `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false"`
	Accept string   `header:"Accept" default:"application/nostr+json;q=0.9,application/x-realy-event;q=0.1" doc:"event encoding format that is expected, priority using mimetype;q=0.x will indicate preference when multiple are available"`
	Body   []string `doc:"list of event Ids"`
}

type EventsOutput struct {
	Limit int `header:"X-Limit" default:"1000" doc:"informs client maximum number of events that they can request"`
}

// RegisterEvents is the implementation of the HTTP API for Events.
func (x *Operations) RegisterEvents(api huma.API) {
	name := "Events"
	description := "Returns the full events from a list of event Ids as a line structured JSON. Auth required to fetch more than 1000 events, and if not enabled, 1000 is the limit."
	path := x.path + "/events"
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
	}, func(ctx context.T, input *EventsInput) (output *struct{}, err error) {
		// log.I.S(input)
		if len(input.Body) == 10000 {
			err = huma.Error400BadRequest(
				"cannot process more than 10000 events in a request")
			return

		}
		var authrequired bool
		if len(input.Body) > 1000 || x.Server.AuthRequired() {
			authrequired = true
		}
		limit := 1000
		if !authrequired {
			limit = 10000
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
			x.Server.Lock()
			// we only allow the first level of the allowed users this kind of access
			if x.Server.OwnersFollowed(string(pubkey)) {
				err = huma.Error403Forbidden(
					fmt.Sprintf(
						"authenticated user %0x does not have permission for this request (owners can use export)",
						pubkey))
				return
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
			w := ctx.Value("http-response").(http.ResponseWriter)
			var binary bool
			precedence := header.ParsePrecedenceValues(r.Header.Get("Accept"))
		done:
			for _, v := range precedence {
				switch v.Value {
				case "application/x-realy-event":
					binary = true
					break done
				case "application/nostr+json":
					break done
				default:
					break done
				}
			}
			w.WriteHeader(200)
			w.Header().Set("X-Limit", fmt.Sprint(limit))
			if err = idsWriter.FetchIds(w, x.Context(), tag.New(evIds...), binary); chk.E(err) {
				return
			}
		}
		return
	})
}
