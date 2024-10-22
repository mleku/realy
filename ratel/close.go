package ratel

func (r *T) Close() (err E) {
	chk.E(r.DB.Sync())
	log.I.F("closing database %s", r.Path())
	if err = r.DB.Flatten(4); chk.E(err) {
		return
	}
	log.D.F("database flattened")
	if err = r.seq.Release(); chk.E(err) {
		return
	}
	log.D.F("database released")
	if err = r.DB.Close(); chk.E(err) {
		return
	}
	log.I.F("database closed")
	return
}
