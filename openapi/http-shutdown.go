package openapi

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
)

type ShutdownInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
}

type ShutdownOutput struct{}

func (x *Operations) RegisterShutdown(api huma.API) {
	name := "Shutdown"
	description := "Shutdown relay"
	path := x.path + "/shutdown"
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
		if !x.Server.Configured() {
			err = huma.Error404NotFound("server is not configured")
			return
		}
		r := ctx.Value("http-request").(*http.Request)
		remote := helpers.GetRemoteFromReq(r)
		authed, _ := x.AdminAuth(r, remote)
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
