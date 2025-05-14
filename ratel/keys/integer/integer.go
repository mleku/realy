package integer

import (
	"bytes"
	"encoding/binary"
	"io"

	"golang.org/x/exp/constraints"

	"realy.lol/chk"
	"realy.lol/ratel/keys"
)

const Len = 4

// T is a 32-bit integer number value.
type T struct {
	Val uint32
}

var _ keys.Element = &T{}

func New[V constraints.Integer](val ...V) (m *T) {
	if len(val) == 0 {
		m = new(T)
		return
	}
	m = &T{uint32(val[0])}
	return
}

func NewFrom(b []byte) (s *T) {
	buf := bytes.NewBuffer(b)
	s = &T{}
	s.Read(buf)
	return
}

func (s *T) Write(buf io.Writer) {
	v := make([]byte, Len)
	binary.LittleEndian.PutUint32(v, s.Val)
	buf.Write(v)
}

func (s *T) Read(buf io.Reader) (el keys.Element) {
	v := make([]byte, Len)
	if n, err := buf.Read(v); chk.E(err) || n != Len {
		return nil
	}
	// log.I.T(v)
	s.Val = binary.LittleEndian.Uint32(v)
	return s
}

func (s *T) Len() int { return Len }
