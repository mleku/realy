package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
)

// Export is a HTTP API method to export the entire content of an event store to an admin user.
type Export struct{ *Server }

// NewExport creates a new Export.
func NewExport(s *Server) (ep *Export) {
	return &Export{Server: s}
}

// ExportInput is the parameters for the HTTP API Export method.
type ExportInput struct {
	Auth        string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Accept      string `header:"Accept" default:"application/nostr+jsonl" enum:"application/nostr+jsonl" required:"false"`
	ContentType string `header:"Content-Type" default:"application/nostr+jsonl" enum:"application/nostr+jsonl" required:"false"`
}

// ExportOutput is the return value of Export. It usually will be line structured JSON. In
// future there may be more output formats.
type ExportOutput struct{ RawBody []byte }

// RegisterExport implements the Export HTTP API method.
func (ep *Export) RegisterExport(api huma.API) {
	name := "Export"
	description := "Export all events (only works with NIP-98/JWT capable client, will not work with UI)"
	path := "/export"
	scopes := []string{"admin", "read"}
	method := http.MethodGet
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"admin"},
		Description: generateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *ExportInput) (resp *huma.StreamResponse, err error) {
		r := ctx.Value("http-request").(*http.Request)
		rr := GetRemoteFromReq(r)
		log.I.F("processing export from %s", rr)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		s := ep.Server
		authed, pubkey := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("Not Authorized")
			return
		}
		log.I.F("export of event data requested on admin port from %s pubkey %0x",
			rr, pubkey)
		sto := s.relay.Storage()
		resp = &huma.StreamResponse{
			func(ctx huma.Context) {
				ctx.SetHeader("Content-Type", "application/nostr+jsonl")
				sto.Export(s.Ctx, ctx.BodyWriter())
				if f, ok := ctx.BodyWriter().(http.Flusher); ok {
					f.Flush()
				} else {
					log.W.F("error: unable to flush")
				}
			},
		}
		return
	})
}
