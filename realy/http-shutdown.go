package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
)

type Shutdown struct{ *Server }

func NewShutdown(s *Server) (ep *Shutdown) {
	return &Shutdown{Server: s}
}

type ShutdownInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
}

type ShutdownOutput struct{}

func (x *Shutdown) RegisterShutdown(api huma.API) {
	name := "Shutdown"
	description := "Shutdown Relay"
	path := "/shutdown"
	scopes := []string{"admin"}
	method := http.MethodGet
	huma.Register(api, huma.Operation{
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"admin"},
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *ShutdownInput) (wgh *ShutdownOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		s := x.Server
		authed, _ := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		x.Server.Shutdown()
		return
	})
}
