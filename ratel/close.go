package ratel

import (
	"realy.lol/chk"
	"realy.lol/log"
)

// Close the database. If the Flatten flag was set, then trigger the flattening of tables before
// shutting down.
func (r *T) Close() (err error) {
	// chk.E(r.DB.Sync())
	// r.WG.Wait()
	log.I.F("closing database %s", r.Path())
	if r.Flatten {
		if err = r.DB.Flatten(4); chk.E(err) {
		}
		log.D.F("database flattened")
	}
	if err = r.seq.Release(); chk.E(err) {
	}
	log.D.F("database released")
	if err = r.DB.Close(); chk.E(err) {
	}
	log.I.F("database closed")
	return
}
