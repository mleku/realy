// Package lang implements a keys.Element for an ISO-639-2 3 letter language code.
package lang

import (
	"io"

	"realy.lol/chk"
	"realy.lol/ratel/keys"
)

const Len = 3

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

// New creates a new kinder.T for reading/writing kind.T values.
func New[V string | []byte](c V) (p *T) {
	if len(c) != Len {
		return &T{Val: make([]byte, Len)}
	}
	return &T{Val: []byte(c)}
}

func (c *T) Write(buf io.Writer) {
	buf.Write(c.Val)
}

func (c *T) Read(buf io.Reader) (el keys.Element) {
	c.Val = make([]byte, Len)
	if n, err := buf.Read(c.Val); chk.E(err) || n != Len {
		return nil
	}
	return c
}

func (c *T) Len() int { return Len }
