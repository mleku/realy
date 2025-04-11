package realy

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/realy/interfaces"
)

type Shutdown struct{ interfaces.Server }

func NewShutdown(s interfaces.Server) (x *Shutdown) {
	return &Shutdown{Server: s}
}

type ShutdownInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
}

type ShutdownOutput struct{}

func (x *Shutdown) RegisterShutdown(api huma.API) {
	name := "Shutdown"
	description := "Shutdown relay"
	path := "/shutdown"
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
	}, func(ctx context.T, input *ShutdownInput) (wgh *ShutdownOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		authed, _ := x.AdminAuth(r)
		if !authed {
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		go func() {
			time.Sleep(time.Second)
			x.Shutdown()
		}()

		return
	})
}
