package api

type Method interface {
	// Handle is a http handler with request and response writer parameters.
	Handle(h H)
	// API returns the capabilities string for a requested codec on the Method
	// implementation.
	API(accept string) (s string)
	// Path returns the API path this API responds to.
	Path() (s string)
}
