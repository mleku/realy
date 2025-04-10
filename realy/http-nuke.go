package realy

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/context"
	"realy.mleku.dev/store"
)

// Nuke is the HTTP API method to wipe the event store of a relay.
type Nuke struct{ *Server }

// NewNuke creates a new Nuke.
func NewNuke(s *Server) (ep *Nuke) { return &Nuke{Server: s} }

// NukeInput is the parameters for the HTTP API method nuke. Note that it has a confirmation
// header that must be provided to prevent accidental invocation of this method.
type NukeInput struct {
	Auth    string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Confirm string `header:"X-Confirm" doc:"must put 'Yes I Am Sure' in this field as confirmation"`
}

// NukeOutput is basically nothing, a 200 or 204 HTTP status response is normal.
type NukeOutput struct{}

// RegisterNuke is the implementation of the Nuke HTTP API method.
func (ep *Nuke) RegisterNuke(api huma.API) {
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
		Description:   generateDescription(description, scopes),
		Security:      []map[string][]string{{"auth": scopes}},
		DefaultStatus: 204,
	}, func(ctx context.T, input *NukeInput) (wgh *NukeOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		rr := GetRemoteFromReq(r)
		s := ep.Server
		authed, pubkey := s.authAdmin(r)
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
		sto := s.relay.Storage()
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
