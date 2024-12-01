package index

import (
	"bytes"
	"fmt"

	"realy.lol/ratel/keys"
)

const Len = 1

type T struct {
	Val by
}

var _ keys.Element = &T{}

func New[V byte | P | no](code ...V) (p *T) {
	var cod by
	switch len(code) {
	case 0:
		cod = by{0}
	default:
		cod = by{byte(code[0])}
	}
	return &T{Val: cod}
}

func Empty() (p *T) {
	return &T{Val: by{0}}
}

func (p *T) Write(buf *bytes.Buffer) {
	if len(p.Val) != Len {
		panic(fmt.Sprintln("must use New or initialize Val with len", Len))
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf *bytes.Buffer) (el keys.Element) {
	p.Val = make(by, Len)
	if n, err := buf.Read(p.Val); chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() no { return Len }
