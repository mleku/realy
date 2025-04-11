package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
)

// DisconnectInput is the parameters for triggering the disconnection of all open websockets.
type DisconnectInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
}

// DisconnectOutput is the result type for the Disconnect HTTP API method.
type DisconnectOutput struct{}

// RegisterDisconnect is the implementation of the HTTP API Disconnect method.
func (x *Operations) RegisterDisconnect(api huma.API) {
	name := "Disconnect"
	description := "Close all open nip-01 websockets"
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
