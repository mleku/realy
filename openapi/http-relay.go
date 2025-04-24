package openapi

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/httpauth"
	"realy.mleku.dev/log"
	"realy.mleku.dev/publish"
	"realy.mleku.dev/realy/helpers"
)

// RelayInput is the parameters for the Event HTTP API method.
type RelayInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false"`
	RawBody []byte
}

// RelayOutput is the return parameters for the HTTP API Relay method.
type RelayOutput struct{ Body string }

// RegisterRelay is the implementatino of the HTTP API Relay method.
func (x *Operations) RegisterRelay(api huma.API) {
	name := "Relay"
	description := "relay an event, don't store it"
	path := x.path + "/relay"
	scopes := []string{"user"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *RelayInput) (output *RelayOutput, err error) {
		if !x.Server.Configured() {
			err = huma.Error503ServiceUnavailable("server is not configured")
			return
		}
		log.I.S(input)
		r := ctx.Value("http-request").(*http.Request)
		remote := helpers.GetRemoteFromReq(r)
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
		var ok bool
		// if there was auth, or no auth, check the relay policy allows accepting the
		// event (no auth with auth required or auth not valid for action can apply
		// here).
		ev := &event.T{}
		if _, err = ev.Unmarshal(input.RawBody); chk.E(err) {
			err = huma.Error406NotAcceptable(err.Error())
			return
		}
		accept, notice, _ := x.AcceptEvent(ctx, ev, r, pubkey, remote)
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
		var authRequired bool

		authRequired = x.Server.AuthRequired()

		publish.P.Deliver(authRequired, x.PublicReadable(), ev)
		return
	})
}
