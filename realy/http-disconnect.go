package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/realy/interfaces"
)

// Disconnect is the HTTP API ta trigger disconnecting all currently open websockets.
type Disconnect struct{ interfaces.Server }

// NewDisconnect creates a new Disconnect.
func NewDisconnect(s interfaces.Server) (ep *Disconnect) {
	return &Disconnect{Server: s}
}

// DisconnectInput is the parameters for triggering the disconnection of all open websockets.
type DisconnectInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
}

// DisconnectOutput is the result type for the Disconnect HTTP API method.
type DisconnectOutput struct{}

// RegisterDisconnect is the implementation of the HTTP API Disconnect method.
func (x *Disconnect) RegisterDisconnect(api huma.API) {
	name := "Disconnect"
	description := "Close all open sockets"
	path := "/disconnect"
	scopes := []string{"admin"}
	method := http.MethodGet
	huma.Register(api, huma.Operation{
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"admin"},
		Description:   helpers.GenerateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *DisconnectInput) (wgh *DisconnectOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		authed, _ := x.AdminAuth(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		x.Disconnect()
		return
	})
}
