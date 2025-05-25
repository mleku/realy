package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/log"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/sha256"
)

type Event struct {
	ser *serial.T
	ev  *event.T
}

// Rescan regenerates all indexes of events to add new indexes in a new version.
func (r *T) Rescan() (err error) {
	r.WG.Add(1)
	defer r.WG.Done()
	evChan := make(chan *Event)
	go func() {
		var count int
		for {
			select {
			case <-r.Ctx.Done():
				log.I.F("completed rescanning %d events", count)
				return
			case e := <-evChan:
				if e == nil {
					log.I.F("completed rescanning %d events", count)
					return
				}
				if err = r.GenerateFulltextIndex(e.ev, e.ser); chk.E(err) {
					return
				}
				if err = r.GenerateLanguageIndex(e.ev, e.ser); chk.E(err) {
					return
				}
			retry:
				if err = r.Update(func(txn *badger.Txn) (err error) {
					// rewrite the indexes
					var indexKeys [][]byte
					indexKeys = GetIndexKeysForEvent(e.ev, e.ser)
					for _, k := range indexKeys {
						var val []byte
						// if k[0] == prefixes.Counter.B() {
						// 	val = keys.Write(createdat.New(timestamp.Now()))
						// }
						if err = txn.Set(k, val); chk.E(err) {
							return
						}
					}
					count++
					// if count%100 == 0 {
					log.I.F("rescanned %d events", count)
					// }
					return
				}); chk.E(err) {
					goto retry
				}
			}
		}
	}()
	err = r.View(func(txn *badger.Txn) (err error) {
		prf := []byte{prefixes.Event.B()}
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(prf); it.Next() {
			item := it.Item()
			if it.Item().ValueSize() == sha256.Size {
				continue
			}
			k := item.KeyCopy(nil)
			ser := serial.New(k[1:])
			var val []byte
			if val, err = item.ValueCopy(nil); chk.E(err) {
				continue
			}
			ev := &event.T{}
			if _, err = r.Unmarshal(ev, val); chk.E(err) {
				return
			}
			evChan <- &Event{ser: ser, ev: ev}
		}
		return
	})
	return err
}
