package context

import (
	"bytes"
	"context"
)

type (
	bo = bool
	by = []byte
	st = string
	er = error
	no = int
	cx = context.Context
)

var (
	equals = bytes.Equal
)
