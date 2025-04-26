// Package pubkey implements an 8 byte truncated public key implementation of a
// keys.Element.
package pubkey

import (
	"fmt"
	"io"

	"realy.lol/chk"
	"realy.lol/ec/schnorr"
	"realy.lol/log"
	"realy.lol/ratel/keys"
)

const Len = 8

type T struct {
	Val []byte
}

var _ keys.Element = &T{}

// New creates a new pubkey prefix, if parameter is omitted, new one is
// allocated (for read) if more than one is given, only the first is used, and
// if the first one is not the correct hexadecimal length of 64, return error.
func New(pk ...[]byte) (p *T, err error) {
	if len(pk) < 1 {
		// allows init with no parameter
		return &T{make([]byte, Len)}, nil
	}
	// // only the first pubkey will be used
	if len(pk[0]) != schnorr.PubKeyBytesLen {
		err = log.E.Err("pubkey hex must be 32 chars, got", len(pk[0]))
		return
	}
	return &T{Val: pk[0][:Len]}, nil
}

func NewFromBytes(pkb []byte) (p *T, err error) {
	if len(pkb) != schnorr.PubKeyBytesLen {
		err = log.E.Err("provided key not correct length, got %d expected %d",
			len(pkb), schnorr.PubKeyBytesLen)
		log.T.S(pkb)
		return
	}
	b := make([]byte, Len)
	copy(b, pkb[:Len])
	p = &T{Val: b}
	return
}

func (p *T) Write(buf io.Writer) {
	if p == nil {
		panic("nil pubkey")
	}
	if p.Val == nil || len(p.Val) != Len {
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
