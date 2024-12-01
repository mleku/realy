package keys

import (
	"bytes"

	"realy.lol/context"
	"realy.lol/lol"
)

type (
	by = []byte
	st = string
	er = error
	no = int
	cx = context.T
)

var (
	log, chk, errorf = lol.Main.Log, lol.Main.Check, lol.Main.Errorf
	equals           = bytes.Equal
)
