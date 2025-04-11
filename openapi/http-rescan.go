package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/store"
)

type RescanInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
}

type RescanOutput struct{}

func (x *Operations) RegisterRescan(api huma.API) {
	name := "Rescan"
	description := "Rescan all events and rewrite their indexes (to enable new indexes on old events)"
	path := "/rescan"
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
	}, func(ctx context.T, input *RescanInput) (wgh *RescanOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		rr := helpers.GetRemoteFromReq(r)
		authed, pubkey := x.AdminAuth(r)
		if !authed {
			err = huma.Error401Unauthorized("not authorized")
			return
		}
		log.I.F("index rescan requested on admin port from %s pubkey %0x",
			rr, pubkey)
		sto := x.Storage()
		if rescanner, ok := sto.(store.Rescanner); ok {
			log.I.F("rescanning")
			if err = rescanner.Rescan(); chk.E(err) {
				return
			}
		}
		return
	})
}
