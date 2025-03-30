// Package fullpubkey implements a keys.Element for a complete 32 byte nostr
// pubkeys.
package fullpubkey

import (
	"bytes"
	"fmt"

	"realy.lol/ec/schnorr"
	"realy.lol/ratel/keys"
)

const Len = schnorr.PubKeyBytesLen

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

func New(evID ...[]byte) (p *T) {
	if len(evID) < 1 || len(evID[0]) < 1 {
		return &T{make([]byte, Len)}
	}
	return &T{Val: evID[0]}
}

func (p *T) Write(buf *bytes.Buffer) {
	if len(p.Val) != Len {
		panic(fmt.Sprintln("must use New or initialize Val with len", Len))
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf *bytes.Buffer) (el keys.Element) {
	// allow uninitialized struct
	if len(p.Val) != Len {
		p.Val = make([]byte, Len)
	}
	if n, err := buf.Read(p.Val); chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() int { return len(p.Val) }
