package tcpkeepalive

import (
	"bytes"

	"realy.mleku.dev/context"
	"realy.mleku.dev/lol"
)

type (
	bo = bool
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
