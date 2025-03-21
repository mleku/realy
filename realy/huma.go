package realy

import (
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// ExposeMiddleware adds the http.Request and http.ResponseWriter to the context
// for the Operations handler.
func ExposeMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Unwrap the request and response objects.
	r, w := humago.Unwrap(ctx)
	ctx = huma.WithValue(ctx, "http-request", r)
	ctx = huma.WithValue(ctx, "http-response", w)
	next(ctx)
}

func NewHuma(router *ServeMux, name, version, description string) (api huma.API) {
	apiConfig := huma.DefaultConfig(name, version)
	apiConfig.Info.Description = description
	// apiConfig.Security = []map[string][]string{{"auth": {}}}
	// apiConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{"auth": {Type: "http", Scheme: "apiKey"}} // apiKey todo:
	apiConfig.DocsPath = "/api"
	api = humago.New(router, apiConfig)
	api.UseMiddleware(ExposeMiddleware)
	return
}

func generateDescription(text string, scopes []string) string {
	if len(scopes) == 0 {
		return text
	}
	result := make([]string, 0)
	for _, value := range scopes {
		result = append(result, "`"+value+"`")
	}
	return text + "<br/><br/>**Scopes**<br/>" + strings.Join(result, ", ")
}
