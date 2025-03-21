package realy

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
)

type Export struct{ *Server }

func NewExport(s *Server) (ep *Export) {
	return &Export{Server: s}
}

type ExportInput struct {
	Auth string `header:"Authorization"`
}

type ExportOutput struct{}

func (ep *EventPost) RegisterExport(api huma.API) {
	name := "Export"
	description := "Export all events"
	path := "/export"
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
	}, func(ctx context.T, input *ExportInput) (wgh *ExportOutput, err error) {
		log.I.S(ctx)
		r := ctx.Value("http-request").(*http.Request)
		w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		log.I.S(r.RemoteAddr, rr)

		log.I.F("export of event data requested on admin port")
		s := ep.Server
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
