package float

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"realy.lol/chk"
	"realy.lol/ratel/keys"
)

const Len = 8

// S is a real (floating point) number value.
type S struct {
	Val float64
}

var _ keys.Element = &S{}

func New(val ...float64) (m *S) {
	if len(val) == 0 {
		m = new(S)
		return
	}
	m = &S{val[0]}
	return
}

func NewFrom(b []byte) (s *S) {
	buf := bytes.NewBuffer(b)
	s = &S{}
	s.Read(buf)
	return
}

func (s *S) Write(buf io.Writer) {
	v := make([]byte, Len)
	binary.BigEndian.PutUint64(v, math.Float64bits(s.Val))
	buf.Write(v)
}

func (s *S) Read(buf io.Reader) (el keys.Element) {
	v := make([]byte, Len)
	if n, err := buf.Read(v); chk.E(err) || n != Len {
		return nil
	}
	// log.I.S(v)
	s.Val = math.Float64frombits(binary.BigEndian.Uint64(v))
	return s
}

func (s *S) Len() int { return Len }
