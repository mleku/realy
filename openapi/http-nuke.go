package openapi

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/store"
)

// NukeInput is the parameters for the HTTP API method nuke. Note that it has a confirmation
// header that must be provided to prevent accidental invocation of this method.
type NukeInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
	Confirm string `header:"X-Confirm" doc:"must put 'Yes I Am Sure' in this field as confirmation"`
}

// NukeOutput is basically nothing, a 200 or 204 HTTP status response is normal.
type NukeOutput struct{}

// RegisterNuke is the implementation of the Nuke HTTP API method.
func (x *Operations) RegisterNuke(api huma.API) {
	name := "Nuke"
	description := "Nuke all events in the database"
	path := "/nuke"
	scopes := []string{"admin", "write"}
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
	}, func(ctx context.T, input *NukeInput) (wgh *NukeOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := helpers.GetRemoteFromReq(r)
		authed, pubkey := x.AdminAuth(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("user not authorized for action")
			return
		}
		if input.Confirm != "Yes I Am Sure" {
			err = huma.Error403Forbidden("Confirm missing or incorrect")
			return
		}
		log.I.F("database nuke request from %s pubkey %0x",
			rr, pubkey)
		sto := x.Storage()
		if nuke, ok := sto.(store.Nukener); ok {
			log.I.F("rescanning")
			if err = nuke.Nuke(); chk.E(err) {
				if strings.HasPrefix(err.Error(), "Value log GC attempt") {
					err = nil
				}
				return
			}
		}
		return
	})
}
