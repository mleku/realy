// Package atag implements a special, optimized handling for keeping a tags
// (address) in a more memory efficient form while working with these tags.
package atag

import (
	"bytes"

	"github.com/davecgh/go-spew/spew"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/log"
)

// T is a data structure for what is found in an `a` tag: kind:pubkey:arbitrary data
type T struct {
	Kind   *kind.T
	PubKey []byte
	DTag   []byte
}

// Marshal an atag.T into raw bytes.
func (t T) Marshal(dst []byte) (b []byte) {
	if t.Kind == nil {
		log.I.F("atag: Kind cannot be nil for Marshal: %s", spew.Sdump(t))
		return
	}
	b = t.Kind.Marshal(dst)
	b = append(b, ':')
	b = hex.EncAppend(b, t.PubKey)
	b = append(b, ':')
	b = append(b, t.DTag...)
	return
}

// Unmarshal an atag.T from its ascii encoding.
func (t *T) Unmarshal(b []byte) (r []byte, err error) {
	split := bytes.Split(b, []byte{':'})
	if len(split) != 3 {
		err = errorf.E("atag: invalid format, expected 3 parts separated by ':' but got %d", len(split))
		return
	}
	// kind
	kin := ints.New(uint16(0))
	if _, err = kin.Unmarshal(split[0]); chk.E(err) {
		return
	}
	t.Kind = kind.New(kin.Uint16())
	// pubkey
	if t.PubKey, err = hex.DecAppend(t.PubKey, split[1]); chk.E(err) {
		return
	}
	// d-tag
	t.DTag = split[2]
	return
}
