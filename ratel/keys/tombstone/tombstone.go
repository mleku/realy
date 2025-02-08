package tombstone

import (
	"bytes"

	"realy.lol/eventid"
	"realy.lol/ratel/keys"
)

const Len = 16

type T struct {
	val []byte
}

var _ keys.Element = &T{}

func Make(eid *eventid.T) (v []byte) {
	v = make([]byte, Len)
	copy(v, eid.Bytes())
	return
}

func New() (t *T) { return new(T) }

func NewWith(eid *eventid.T) (t *T) {
	t = &T{val: Make(eid)}
	return
}

func (t *T) Write(buf *bytes.Buffer) {
	buf.Write(t.val)
}

func (t *T) Read(buf *bytes.Buffer) (el keys.Element) {
	b := make([]byte, Len)
	if n, err := buf.Read(b); chk.E(err) || n < Len {
		log.I.S(n, err)
		return nil
	}
	t.val = b
	return &T{val: b}
}

func (t *T) Len() int { return Len }
