package atag

import (
	"bytes"

	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
)

type T struct {
	Kind   *kind.T
	PubKey []byte
	DTag   []byte
}

func (t T) Marshal(dst []byte) (b []byte) {
	b = t.Kind.Marshal(dst)
	b = append(b, ':')
	b = hex.EncAppend(b, t.PubKey)
	b = append(b, ':')
	b = append(b, t.DTag...)
	return
}

func (t *T) Unmarshal(b []byte) (r []byte, err error) {
	split := bytes.Split(b, []byte{':'})
	if len(split) != 3 {
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
