package realy

import (
	"io"
)

type I interface {
	Label() string
	Write(w io.Writer) (err E)
	JSON
}
