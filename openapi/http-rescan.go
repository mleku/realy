package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/log"
	"realy.lol/realy/helpers"
	"realy.lol/store"
)

type RescanInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
}

type RescanOutput struct{}

func (x *Operations) RegisterRescan(api huma.API) {
	name := "Rescan"
	description := "Rescan all events and rewrite their indexes (to enable new indexes on old events)"
	path := x.path + "/rescan"
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
		remote := helpers.GetRemoteFromReq(r)
		authed, pubkey := x.AdminAuth(r, remote)
		if !authed {
			err = huma.Error401Unauthorized("not authorized")
			return
		}
		log.I.F("index rescan requested on admin port from %s pubkey %0x",
			remote, pubkey)
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
