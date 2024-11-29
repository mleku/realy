package config

import (
	"bytes"

	"realy.lol/context"
	"realy.lol/lol"
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
