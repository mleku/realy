package ratel

import (
	"mleku.dev/ratel/keys/index"
)

func (r *T) Nuke() (err E) {
	log.W.F("nukening database at %s", r.dataDir)
	if err = r.DB.DropPrefix([][]byte{
		{index.Event.B()},
		{index.CreatedAt.B()},
		{index.Id.B()},
		{index.Kind.B()},
		{index.Pubkey.B()},
		{index.PubkeyKind.B()},
		{index.Tag.B()},
		{index.Tag32.B()},
		{index.TagAddr.B()},
		{index.Counter.B()},
	}...); chk.E(err) {
		return
	}
	if err = r.DB.RunValueLogGC(0.8); chk.E(err) {
		return
	}
	return
}
