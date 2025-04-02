package realy

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/context"
	"realy.lol/store"
)

type Configuration struct{ *Server }

func NewConfiguration(s *Server) (ep *Configuration) {
	return &Configuration{Server: s}
}

type ConfigurationInput struct {
	Auth string               `header:"Authorization" doc:"nostr nip-98 or JWT token for authentication" required:"true" example:"Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJFUzI1N2ZGFkNjZlNDdkYjJmIiwic3ViIjoiaHR0cDovLzEyNy4wLjAuMSJ9.cHT_pB3wTLxUNOqxYL6fxAYUJXNKBXcOnYLlkO1nwa7BHr9pOTQzNywJpc3MM2I0N2UziOiI0YzgwMDI1N2E1ODhhODI4NDlkMDIsImV4cCIQ5ODE3YzJiZGFhZDk4NGMgYtGi6MTc0Mjg40NWFkOWYCzvHyiXtIyNWEVZiaWF0IjoxNzQyNjMwMjM3LClZPtt0w_dJxEpYcSIEcY4wg"`
	Body *store.Configuration `doc:"the new configuration"`
}

type ConfigurationOutput struct {
	Body *store.Configuration `doc:"the current configuration"`
}

func (x *Configuration) RegisterConfigurationSet(api huma.API) {
	name := "ConfigurationSet"
	description := "set the current configuration"
	path := "/configuration/set"
	scopes := []string{"admin", "write"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"admin"},
		Description: generateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *ConfigurationInput) (wgh *struct{}, err error) {
		log.I.S(input)
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		s := x.Server
		authed, _ := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		sto := s.relay.Storage()
		if c, ok := sto.(store.Configurationer); ok {
			if err = c.SetConfiguration(input.Body); chk.E(err) {
				return
			}
			s.Configuration = input.Body
		}
		return
	})
}

func (x *Configuration) RegisterConfigurationGet(api huma.API) {
	name := "ConfigurationGet"
	description := "fetch the current configuration"
	path := "/configuration/get"
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
	}, func(ctx context.T, input *ConfigurationInput) (output *ConfigurationOutput, err error) {
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		s := x.Server
		authed, _ := s.authAdmin(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		// sto := s.relay.Storage()
		// if c, ok := sto.(store.Configurationer); ok {
		// 	var cfg *store.Configuration
		// 	if cfg, err = c.GetConfiguration(); chk.E(err) {
		// 		return
		// 	}
		output = &ConfigurationOutput{Body: s.Configuration}
		// }
		return
	})
}
