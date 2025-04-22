package ratel

import (
	"realy.mleku.dev/chk"
	"realy.mleku.dev/log"
	"realy.mleku.dev/ratel/prefixes"
)

func (r *T) Nuke() (err error) {
	log.W.F("nuking database at %s", r.dataDir)
	log.I.S(prefixes.AllPrefixes)
	if err = r.DB.DropPrefix(prefixes.AllPrefixes...); chk.E(err) {
		return
	}
	if err = r.DB.RunValueLogGC(0.8); chk.E(err) {
		return
	}
	return
}
