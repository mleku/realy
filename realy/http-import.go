package realy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
)

type Import struct{ *Server }

func NewImport(s *Server) (ep *Import) {
	return &Import{Server: s}
}

type ImportInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	RawBody []byte
}

type ImportOutput struct{}

func (ep *Import) RegisterImport(api huma.API) {
	name := "Import"
	description := "Import events from line structured JSON (jsonl)"
	path := "/import"
	scopes := []string{"admin", "write"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"admin"},
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *ImportInput) (wgh *ImportOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		authed, pubkey := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized(
				fmt.Sprintf("invalid: %s", err.Error()))
			return
		}
		sto := s.relay.Storage()
		if len(input.RawBody) > 0 {
			read := bytes.NewBuffer(input.RawBody)
			sto.Import(read)
			if realy, ok := s.relay.(*app.Relay); ok {
				realy.ZeroLists()
				realy.CheckOwnerLists(context.Bg())
			}
		} else {
			log.I.F("import of event data requested on admin port from %s pubkey %0x", rr, pubkey)
			read := io.LimitReader(r.Body, r.ContentLength)
			sto.Import(read)
			if realy, ok := s.relay.(*app.Relay); ok {
				realy.ZeroLists()
				realy.CheckOwnerLists(context.Bg())
			}
		}
		return
	})
}
