package ratel

import (
	"realy.lol/ratel/prefixes"
)

func (r *T) Nuke() (err error) {
	log.W.F("nuking database at %s", r.dataDir)
	if err = r.DB.DropPrefix([][]byte{
		{prefixes.Event.B()},
		{prefixes.CreatedAt.B()},
		{prefixes.Id.B()},
		{prefixes.Kind.B()},
		{prefixes.Pubkey.B()},
		{prefixes.PubkeyKind.B()},
		{prefixes.Tag.B()},
		{prefixes.Tag32.B()},
		{prefixes.TagAddr.B()},
		{prefixes.Counter.B()},
	}...); chk.E(err) {
		return
	}
	if err = r.DB.RunValueLogGC(0.8); chk.E(err) {
		return
	}
	return
}
