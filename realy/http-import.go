package realy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/cmd/realy/app"
	"realy.mleku.dev/context"
)

// Import is a HTTP API method that accepts events as minified, line structured JSON.
type Import struct{ *Server }

// NewImport creates a new Import.
func NewImport(s *Server) (ep *Import) {
	return &Import{Server: s}
}

// ImportInput is the parameters of an import operation, authentication and the stream of line
// structured JSON events.
type ImportInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	RawBody []byte
}

// ImportOutput is nothing, basically, a 204 or 200 status is expected.
type ImportOutput struct{}

// RegisterImport is the implementation of the Import operation.
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
		authed, pubkey := s.authAdmin(r, time.Minute*10)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized(
				fmt.Sprintf("user %0x not authorized for action", pubkey))
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
			log.I.F("import of event data requested on admin port from %s pubkey %0x", rr,
				pubkey)
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
