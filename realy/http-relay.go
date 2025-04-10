package realy

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/httpauth"
	"realy.mleku.dev/relay"
)

// Relay is the HTTP API method for submitting an event only to subscribers and not saving it.
type Relay struct{ *Server }

// NewRelay creates a new Relay.
func NewRelay(s *Server) (ep *Relay) {
	return &Relay{Server: s}
}

// RelayInput is the parameters for the Event HTTP API method.
type RelayInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"false" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	RawBody []byte
}

// RelayOutput is the return parameters for the HTTP API Relay method.
type RelayOutput struct{ Body string }

// RegisterRelay is the implementatino of the HTTP API Relay method.
func (ep *Relay) RegisterRelay(api huma.API) {
	name := "Relay"
	description := "Relay an event, don't store it"
	path := "/relay"
	scopes := []string{"user"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: generateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *RelayInput) (output *RelayOutput, err error) {
		log.I.S(input)
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		var valid bool
		var pubkey []byte
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
		var ok bool
		// if there was auth, or no auth, check the relay policy allows accepting the
		// event (no auth with auth required or auth not valid for action can apply
		// here).
		ev := &event.T{}
		if _, err = ev.Unmarshal(input.RawBody); chk.E(err) {
			err = huma.Error406NotAcceptable(err.Error())
			return
		}
		accept, notice, _ := s.relay.AcceptEvent(ctx, ev, r, rr, pubkey)
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
		var ar relay.Authenticator
		if ar, ok = s.relay.(relay.Authenticator); ok {
			authRequired = ar.AuthEnabled()
		}
		s.Listeners.NotifySubscribers(authRequired, s.publicReadable, ev)
		// s := ep.Server
		return
	})
}
