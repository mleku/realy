package kinds

import (
	"bytes"

	"mleku.dev/context"
	"mleku.dev/lol"
)

type (
	B   = []byte
	S   = string
	E   = error
	N   = int
	Ctx = context.T
)

var (
	log, chk, errorf = lol.Main.Log, lol.Main.Check, lol.Main.Errorf
	equals           = bytes.Equal
)
