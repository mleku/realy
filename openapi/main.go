package openapi

import (
	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/servemux"
)

type Operations struct {
	interfaces.Server
	path string
	*servemux.S
}

// New creates a new openapi.Operations and registers its methods.
func New(s interfaces.Server, name, version, description string, path string,
	sm *servemux.S) {

	a := NewHuma(sm, name, version, description)
	huma.AutoRegister(a, &Operations{Server: s, path: path})
	return
}
