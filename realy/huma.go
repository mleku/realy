package realy

import (
	"net/http"
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
	config := huma.DefaultConfig(name, version)
	config.Info.Description = description
	// config.Security = []map[string][]string{{"auth": {}}}
	// config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{"auth": {Type: "http", Scheme: "apiKey"}} // apiKey todo:
	// config.DocsPath = "/api"
	config.DocsPath = ""
	router.ServeMux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
  <head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`))
	})

	api = humago.New(router, config)
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
