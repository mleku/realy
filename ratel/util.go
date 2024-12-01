package ratel

import (
	"bytes"

	"realy.lol/context"
	"realy.lol/lol"
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
