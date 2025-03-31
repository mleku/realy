// Package examples is an embeded jsonl format of a collection of events
// intended to be used to test an event codec.
package examples

import (
	_ "embed"
)

// todo: re-encode this stuff as binary events with compression

//go:embed out.jsonl
var Cache []byte
