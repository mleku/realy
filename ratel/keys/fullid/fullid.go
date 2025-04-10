// Package fullid implements a keys.Element for a complete 32 byte event Ids.
package fullid

import (
	"fmt"
	"io"

	"realy.mleku.dev/ratel/keys"
	"realy.mleku.dev/sha256"

	"realy.mleku.dev/eventid"
)

const Len = sha256.Size

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

func New(evID ...*eventid.T) (p *T) {
	if len(evID) < 1 {
		return &T{make([]byte, Len)}
	}
	return &T{Val: evID[0].Bytes()}
}

func (p *T) Write(buf io.Writer) {
	if len(p.Val) != Len {
		panic(fmt.Sprintln("must use New or initialize Val with len", Len))
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf io.Reader) (el keys.Element) {
	// allow uninitialized struct
	if len(p.Val) != Len {
		p.Val = make([]byte, Len)
	}
	if n, err := buf.Read(p.Val); chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() int { return Len }
