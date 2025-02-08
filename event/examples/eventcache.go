package examples

import (
	_ "embed"
)

// todo: re-encode this stuff as binary events with compression

//go:embed out.jsonl
var Cache []byte
