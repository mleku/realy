package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/log"
	"realy.mleku.dev/realy/helpers"
	"realy.mleku.dev/store"
)

// ConfigurationSetInput is the parameters for HTTP API method to set Configuration.
type ConfigurationSetInput struct {
	Auth string               `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
	Body *store.Configuration `doc:"the new configuration"`
}

// ConfigurationGetInput is the parameters for HTTP API method to get Configuration.
type ConfigurationGetInput struct {
	Auth   string `header:"Authorization" doc:"nostr nip-98 (and expiring variant)" required:"true"`
	Accept string `header:"Accept" default:"application/json" enum:"application/json" required:"true"`
}

// ConfigurationGetOutput is the result of getting Configuration.
type ConfigurationGetOutput struct {
	Body store.Configuration `doc:"the current configuration"`
}

// RegisterConfigurationSet implements the HTTP API for setting Configuration.
func (x *Operations) RegisterConfigurationSet(api huma.API) {
	name := "ConfigurationSet"
	description := "Set the configuration"
	path := "/configuration/set"
	scopes := []string{"admin", "write"}
	method := http.MethodPost
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"admin"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *ConfigurationSetInput) (wgh *struct{}, err error) {
		log.I.S(input)
		r := ctx.Value("http-request").(*http.Request)
		// w := ctx.Value("http-response").(http.ResponseWriter)
		// rr := GetRemoteFromReq(r)
		authed, _ := x.AdminAuth(r)
		if !authed {
			// pubkey = ev.Pubkey
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		sto := x.Storage()
		if c, ok := sto.(store.Configurationer); ok {
			if err = c.SetConfiguration(input.Body); chk.E(err) {
				return
			}
			x.SetConfiguration(input.Body)
		}
		return
	})
}

// RegisterConfigurationGet implements the HTTP API for getting the Configuration.
func (x *Operations) RegisterConfigurationGet(api huma.API) {
	name := "ConfigurationGet"
	description := "Fetch the current configuration"
	path := "/configuration/get"
	scopes := []string{"admin", "read"}
	method := http.MethodGet
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"admin"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *ConfigurationGetInput) (output *ConfigurationGetOutput,
		err error) {
		r := ctx.Value("http-request").(*http.Request)
		authed, _ := x.AdminAuth(r)
		if !authed {
			err = huma.Error401Unauthorized("authorization required")
			return
		}
		output = &ConfigurationGetOutput{Body: x.Configuration()}
		// }
		return
	})
}
