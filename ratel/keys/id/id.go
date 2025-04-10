// Package id implements a keys.Element for a truncated event Ids containing the
// first 8 bytes of an eventid.T.
package id

import (
	"fmt"
	"io"
	"strings"

	"realy.mleku.dev/ratel/keys"
	"realy.mleku.dev/sha256"

	"realy.mleku.dev/eventid"
	"realy.mleku.dev/hex"
)

const Len = 8

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

func New(evID ...*eventid.T) (p *T) {
	if len(evID) < 1 || len(evID[0].String()) < 1 {
		return &T{make([]byte, Len)}
	}
	evid := evID[0].String()
	if len(evid) < 64 {
		evid = strings.Repeat("0", 64-len(evid)) + evid
	}
	if len(evid) > 64 {
		evid = evid[:64]
	}
	b, err := hex.Dec(evid[:Len*2])
	if chk.E(err) {
		return
	}
	return &T{Val: b}
}

func NewFromBytes(b []byte) (p *T, err error) {
	if len(b) != sha256.Size {
		err = errorf.E("event Id must be 32 bytes got: %d %0x", len(b), b)
		return
	}
	p = &T{Val: b[:Len]}
	return
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
