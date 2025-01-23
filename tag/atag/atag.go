package atag

import (
	"realy.lol/kind"
	"bytes"
	"realy.lol/ints"
	"realy.lol/hex"
)

type T struct {
	Kind   *kind.T
	PubKey by
	DTag   by
}

func (t T) Marshal(dst by) (b by) {
	b = t.Kind.Marshal(dst)
	b = append(b, ':')
	b = hex.EncAppend(b, t.PubKey)
	b = append(b, ':')
	b = append(b, t.DTag...)
	return
}

func (t *T) Unmarshal(b by) (r by, err er) {
	split := bytes.Split(b, by{':'})
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
