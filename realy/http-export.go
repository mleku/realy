package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
)

type Export struct{ *Server }

func NewExport(s *Server) (ep *Export) {
	return &Export{Server: s}
}

type ExportInput struct {
	Auth string `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
}

type ExportOutput struct{ RawBody []byte }

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
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
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
				ctx.SetHeader("Content-Type", "application/x-jsonl")
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

// func (ep *Export) RegisterExport(api huma.API) {
// 	name := "Export"
// 	description := "Export all events (only works with NIP-98/JWT capable client, will not work with UI), produces standard HTTP SSE format - event: event (new line) data: (wire format json of event) (newline)"
// 	path := "/export"
// 	scopes := []string{"admin", "read"}
// 	method := http.MethodGet
// 	sse.Register(api, huma.Operation{
// 		OperationID:   name,
// 		Summary:       name,
// 		Path:          path,
// 		Method:        method,
// 		Tags:          []string{"admin"},
// 		Description:   generateDescription(description, scopes),
// 		Security:      []map[string][]string{{"auth": scopes}},
// 		DefaultStatus: 204,
// 	}, map[string]any{
// 		// "event": event.J{},
// 	}, func(ctx context.T, input *ExportInput, send sse.Sender) {
// 		s := ep.Server
// 		w := make(chan *event.J)
// 		var err error
// 		go func() {
// 			// start up the sender
// 			for ev := range w {
// 				send(sse.Message{})
// 				if err = send.Data(ev); chk.E(err) {
// 					return
// 				}
// 			}
// 		}()
// 		sto := s.relay.Storage()
// 		if exporter, ok := sto.(store.ConcurrentExporter); ok {
// 			// start the exporter
// 			exporter.ExportConcurrent(s.Ctx, w)
// 		} else {
// 			log.I.F("no concurrent exporter available")
// 			return
// 		}
// 	})
// }
