package realy

import (
	. "nostr.mleku.dev"
)

type Notice struct {
	Kind    S `json:"kind"`
	Message S `json:"message"`
}
