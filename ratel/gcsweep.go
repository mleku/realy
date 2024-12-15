package ratel

import (
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/event"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/ratel/keys/prefixes"
)

func (r *T) GCSweep(evs, idxs DelItems) (err er) {
	// first we must gather all the indexes of the relevant events
	started := time.Now()
	batch := r.DB.NewWriteBatch()
	defer func() {
		log.I.Ln("flushing GC sweep batch")
		if err = batch.Flush(); chk.E(err) {
			return
		}
		if vlerr := r.DB.RunValueLogGC(0.5); vlerr == nil {
			log.I.Ln("value log cleaned up")
		}
		chk.E(r.DB.Sync())
		batch.Cancel()
		log.I.Ln("completed sweep in", time.Now().Sub(started), r.Path())
	}()
	// var wg sync.WaitGroup
	// go func() {
	// 	wg.Add(1)
	// 	defer wg.Done()
	stream := r.DB.NewStream()
	// get all the event indexes to delete/prune
	stream.Prefix = prefixes.Event.Key()
	stream.ChooseKey = func(item *badger.Item) (boo bo) {
		if item.KeySize() != 1+serial.Len {
			return
		}
		if item.IsDeletedOrExpired() {
			return
		}
		key := item.KeyCopy(nil)
		ser := serial.FromKey(key).Uint64()
		var found bo
		for i := range evs {
			if evs[i] == ser {
				found = true
				break
			}
		}
		if !found {
			return
		}
		if r.HasL2 {
			// if it's already pruned, skip
			if item.ValueSize() == sha256.Size {
				return
			}
			// if there is L2 we are only pruning (replacing event with the ID hash)
			var evb by
			if evb, err = item.ValueCopy(nil); chk.E(err) {
				return
			}
			ev := &event.T{}
			var rem by
			if rem, err = r.Unmarshal(ev, evb); chk.E(err) {
				return
			}
			if len(rem) != 0 {
				log.I.S(rem)
			}
			// otherwise we are deleting
			if err = batch.Delete(key); chk.E(err) {
				return
			}
			if err = batch.Set(key, ev.ID); chk.E(err) {
				return
			}
			return
		} else {
			// otherwise we are deleting
			if err = batch.Delete(key); chk.E(err) {
				return
			}
		}
		return
	}
	// execute the event prune/delete
	if err = stream.Orchestrate(r.Ctx); chk.E(err) {
		return
	}
	// }()
	// next delete all the indexes
	if len(idxs) > 0 && r.HasL2 {
		log.I.Ln("pruning indexes")
		// we have to remove everything
		prfs := []by{prefixes.Event.Key()}
		prfs = append(prfs, prefixes.FilterPrefixes...)
		prfs = append(prfs, by{prefixes.Counter.B()})
		for _, prf := range prfs {
			stream = r.DB.NewStream()
			stream.Prefix = prf
			stream.ChooseKey = func(item *badger.Item) (boo bo) {
				if item.IsDeletedOrExpired() || item.KeySize() < serial.Len+1 {
					return
				}
				key := item.KeyCopy(nil)
				ser := serial.FromKey(key).Uint64()
				var found bo
				for _, idx := range idxs {
					if idx == ser {
						found = true
						break
					}
				}
				if !found {
					return
				}
				// log.I.F("deleting index %x %d", prf, ser)
				if err = batch.Delete(key); chk.E(err) {
					return
				}
				return
			}
			if err = stream.Orchestrate(r.Ctx); chk.E(err) {
				return
			}
			log.T.Ln("completed index prefix", prf)
		}
	}
	return
}
