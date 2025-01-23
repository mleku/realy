package arb

import (
	"bytes"

	"realy.lol/ratel/keys"
)

// T is an arbitrary length byte string. In any construction there can only be one with arbitrary length. Custom lengths
// can be created by calling New with the custom length in it, both for Read and Write operations.
type T struct {
	Val by
}

var _ keys.Element = &T{}

// New creates a new arb.T. This must have the expected length for the provided byte slice as this is what the Read
// method will aim to copy. In general this will be a bounded field, either the final or only arbitrary length field in
// a key.
func New[V by | st](s V) (p *T) {
	b := by(s)
	if len(b) == 0 {
		log.T.Ln("empty or nil slice is the same as zero value, " +
			"use keys.ReadWithArbElem")
		return &T{}
	}
	return &T{Val: b}
}

func NewWithLen(l no) (p *T) { return &T{Val: make(by, l)} }

func (p *T) Write(buf *bytes.Buffer) {
	if len(p.Val) == 0 {
		log.T.Ln("empty slice has no effect")
		return
	}
	buf.Write(p.Val)
}

func (p *T) Read(buf *bytes.Buffer) (el keys.Element) {
	if len(p.Val) < 1 {
		log.T.Ln("empty slice has no effect")
		return
	}
	if _, err := buf.Read(p.Val); chk.E(err) {
		return nil
	}
	return p
}

func (p *T) Len() no {
	if p == nil {
		panic("uninitialized pointer to arb.T")
	}
	return len(p.Val)
}

// ReadWithArbElem is a variant of Read that recognises an arbitrary length element by its zero length and imputes its
// actual length by the byte buffer size and the lengths of the fixed length fields.
//
// For reasons of space efficiency, it is not practical to use TLVs for badger database key fields, so this will panic
// if there is more than one arbitrary length element.
func ReadWithArbElem(b by, elems ...keys.Element) {
	var arbEl no
	var arbSet bo
	l := len(b)
	for i, el := range elems {
		elLen := el.Len()
		l -= elLen
		if elLen == 0 {
			if arbSet {
				panic("cannot have more than one arbitrary length field in a key")
			}
			arbEl = i
			arbSet = true
		}
	}
	// now we can say that the remainder is the correct length for the arb element
	elems[arbEl] = New(make(by, l))
	buf := bytes.NewBuffer(b)
	for _, el := range elems {
		el.Read(buf)
	}
}
