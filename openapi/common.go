package openapi

import (
	"realy.mleku.dev/realy/interfaces"
)

type Operations struct{ interfaces.Server }

// NewOperations creates a new openapi.Operations..
func NewOperations(s interfaces.Server) (ep *Operations) {
	return &Operations{Server: s}
}
