package serial

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"realy.lol/ratel/keys"
)

const Len = 8

// T is a badger DB serial number used for conflict free event record keys.
type T struct {
	Val by
}

var _ keys.Element = &T{}

// New returns a new serial record key.Element - if nil or short slice is given,
// initialize a fresh one with Len (for reading), otherwise if equal or longer,
// trim if long and store into struct (for writing).
func New(ser by) (p *T) {
	switch {
	case len(ser) < Len:
		// log.I.Ln("empty serial")
		// allows use of nil to init
		ser = make(by, Len)
	default:
		ser = ser[:Len]
	}
	return &T{Val: ser}
}

// FromKey expects the last Len bytes of the given slice to be the serial.
func FromKey(k by) (p *T) {
	if len(k) < Len {
		panic(fmt.Sprintf("cannot get a serial without at least 8 bytes %x", k))
	}
	key := make(by, Len)
	copy(key, k[len(k)-Len:])
	return &T{Val: key}
}

func Make(s uint64) (ser by) {
	ser = make(by, 8)
	binary.BigEndian.PutUint64(ser, s)
	return
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
		p.Val = make(by, Len)
	}
	if n, err := buf.Read(p.Val); chk.E(err) || n != Len {
		return nil
	}
	return p
}

func (p *T) Len() no            { return Len }
func (p *T) Uint64() (u uint64) { return binary.BigEndian.Uint64(p.Val) }

// Match compares a key bytes to a serial, all indexes have the serial at
// the end indicating the event record they refer to, and if they match returns
// true.
func Match(index, ser by) bo {
	l := len(index)
	if l < Len {
		return false
	}
	return bytes.Compare(index[l-Len:], ser) == 0
}
