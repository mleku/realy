package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
	"realy.lol/store"
)

type Rescan struct{ *Server }

func NewRescan(s *Server) (ep *Rescan) {
	return &Rescan{Server: s}
}

type RescanInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
}

type RescanOutput struct{}

func (ep *Rescan) RegisterRescan(api huma.API) {
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
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *RescanInput) (wgh *RescanOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		authed, pubkey := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("not authorized")
			return
		}
		log.I.F("index rescan requested on admin port from %s pubkey %0x",
			rr, pubkey)
		sto := s.relay.Storage()
		if rescanner, ok := sto.(store.Rescanner); ok {
			log.I.F("rescanning")
			if err = rescanner.Rescan(); chk.E(err) {
				return
			}
		}
		return
	})
}
