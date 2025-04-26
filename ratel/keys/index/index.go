// Package index implements the single byte prefix of the database keys. This
// means a limit of 256 tables but is plenty for a single purpose nostr event
// store.
package index

import (
	"fmt"
	"io"

	"realy.lol/chk"
	"realy.lol/ratel/keys"
)

const Len = 1

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

func New[V byte | P | int](code ...V) (p *T) {
	var cod []byte
	switch len(code) {
	case 0:
		cod = []byte{0}
	default:
		cod = []byte{byte(code[0])}
	}
	return &T{Val: cod}
}

func Empty() (p *T) {
	return &T{Val: []byte{0}}
}

func (p *T) Write(buf io.Writer) {
	if len(p.Val) != Len {
		panic(fmt.Sprintln("must use New or initialize Val with len", Len))
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf io.Reader) (el keys.Element) {
	p.Val = make([]byte, Len)
	if n, err := buf.Read(p.Val); chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() int { return Len }
