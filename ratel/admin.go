package ratel

import (
	"bufio"
	"io"

	"github.com/dgraph-io/badger/v4"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/ratel/keys/index"
)

const maxLen = 500000000

// Import accepts an event
func (r *T) Import(rr io.Reader) {
	r.Flatten = true
	scan := bufio.NewScanner(rr)
	buf := make(B, maxLen)
	scan.Buffer(buf, maxLen)
	var err E
	var count N
	for scan.Scan() {
		b := scan.Bytes()
		// if len(b) > 8192 {
		// 	log.I.F("saving,%s", b)
		// }
		ev := &event.T{}
		if _, err = ev.UnmarshalJSON(b); err != nil {
			continue
		}
		if err = r.SaveEvent(r.Ctx, ev); err != nil {
			continue
		}
		count++
		if count > 0 && count%10000 == 0 {
			chk.T(r.DB.Sync())
			log.I.F("imported 10000/%d events, running GC on new data", count)
			chk.T(r.DB.RunValueLogGC(0.5))
		}
	}
	err = scan.Err()
	if chk.E(err) {
	}
	return
}

func (r *T) Export(c context.T, w io.Writer) {
	var counter int
	err := r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: index.Event.Key()})
		defer it.Close()
		var started bool
		for it.Rewind(); it.Valid(); it.Next() {
			select {
			case <-r.Ctx.Done():
				return
			default:
			}
			item := it.Item()
			if started {
				_, _ = w.Write(B{'\n'})
			} else {
				started = true
			}
			b, e := item.ValueCopy(nil)
			if e != nil {
				err = nil
				continue
			}
			var rem B
			ev := &event.T{}
			if rem, err = ev.UnmarshalBinary(b); chk.E(err) {
				err = nil
				continue
			}
			if len(rem) > 0 {
				log.T.S(rem)
			}
			if _, err = w.Write(ev.Serialize()); chk.E(err) {
				// err = nil
				// continue
				return
			}
			counter++
		}
		return
	})
	chk.E(err)
	log.I.Ln("exported", counter, "events")
	return
}
