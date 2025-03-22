package realy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/httpauth"
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
		OperationID:   name,
		Summary:       name,
		Path:          path,
		Method:        method,
		Tags:          []string{"admin"},
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *ExportInput) (wgh *ExportOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		var valid bool
		var pubkey []byte
		if valid, pubkey, err = httpauth.CheckAuth(r, s.JWTVerifyFunc); chk.E(err) {
			return
		}
		if !valid {
			// pubkey = ev.PubKey
			err = huma.Error401Unauthorized(
				fmt.Sprintf("invalid: %s", err.Error()))
			return
		}
		log.I.F("export of event data requested on admin port from %s pubkey %0x",
			rr, pubkey)
		sto := s.relay.Storage()
		if strings.Count(r.URL.Path, "/") > 1 {
			split := strings.Split(r.URL.Path, "/")
			if len(split) != 3 {
				fprintf(w, "incorrectly formatted export parameter: '%s'", r.URL.Path)
				return
			}
			switch split[2] {
			case "users":
				if rl, ok := s.relay.(*app.Relay); ok {
					follows := make([][]byte, 0, len(rl.Followed))
					for f := range rl.Followed {
						follows = append(follows, []byte(f))
					}
					sto.Export(s.Ctx, w, follows...)
				}
			default:
				var exportPubkeys [][]byte
				pubkeys := strings.Split(split[2], "-")
				for _, pubkey := range pubkeys {
					var pk []byte
					pk, err = hex.Dec(pubkey)
					if err != nil {
						log.E.F("invalid public key '%s' in parameters", pubkey)
						continue
					}
					exportPubkeys = append(exportPubkeys, pk)
				}
				sto.Export(s.Ctx, w, exportPubkeys...)
			}
		} else {
			sto.Export(s.Ctx, w)
		}
		return
	})
}
