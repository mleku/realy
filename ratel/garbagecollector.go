package ratel

import (
	"time"

	"realy.lol/units"
)

// GarbageCollector starts up a ticker that runs a check on space utilisation
// and when it exceeds the high-water mark, prunes back to the low-water mark.
//
// This function should be invoked as a goroutine, and will terminate when the
// backend context is canceled.
func (r *T) GarbageCollector() {
	log.D.F("starting badger back-end garbage collector: max size %0.3f MB; "+
		"high water %0.3f MB; "+
		"low water %0.3f MB "+
		"(MB = %d bytes) "+
		"GC check frequency %v %s",
		float32(r.DBSizeLimit)/units.Mb,
		float32(r.DBHighWater*r.DBSizeLimit/100)/units.Mb,
		float32(r.DBLowWater*r.DBSizeLimit/100)/units.Mb,
		units.Mb,
		r.GCFrequency,
		r.Path,
	)
	var err error
	if err = r.GCRun(); chk.E(err) {
	}
	GCticker := time.NewTicker(r.GCFrequency)
	syncTicker := time.NewTicker(r.GCFrequency * 10)
out:
	for {
		select {
		case <-r.Ctx.Done():
			log.W.Ln("stopping event GC ticker")
			break out
		case <-GCticker.C:
			// log.T.Ln("running GC", r.Path)
			if err = r.GCRun(); chk.E(err) {
			}
		case <-syncTicker.C:
			chk.E(r.DB.Sync())
		}
	}
	log.I.Ln("closing badger event store garbage collector")
}

func (r *T) GCRun() (err error) {
	log.T.Ln("running GC", r.Path)
	var pruneEvents, pruneIndexes DelItems
	if pruneEvents, pruneIndexes, err = r.GCMark(); chk.E(err) {
		return
	}
	if len(pruneEvents) < 1 && len(pruneIndexes) < 1 {
		log.I.Ln("GC sweep unnecessary")
		return
	}
	if err = r.GCSweep(pruneEvents, pruneIndexes); chk.E(err) {
		return
	}
	return
}
