package ratel

func (r *T) Close() (err error) {
	// chk.E(r.DB.Sync())
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
