package ratel

import (
	"io"

	"github.com/dgraph-io/badger/v4"
	"realy.lol/event"
	"realy.lol/ratel/keys/index"
)

func (r *T) Import(rr io.Reader) {

	return
}

func (r *T) Export(w io.Writer) {
	var counter int
	err := r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: index.Event.Key()})
		defer it.Close()
		var started bool
		for it.Rewind(); it.Valid(); it.Next() {
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
				err = nil
				continue
			}
			counter++
		}
		return
	})
	chk.E(err)
	log.I.Ln("exported", counter, "events")
	return
}
