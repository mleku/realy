package main

import (
	"bytes"

	"realy.lol/context"
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
	equals = bytes.Equal
)
