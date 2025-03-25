package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/event"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/sha256"
	"realy.lol/timestamp"
)

// Rescan regenerates all indexes of events to add new indexes in a new version.
func (r *T) Rescan() (err error) {
	var evKeys [][]byte
	err = r.View(func(txn *badger.Txn) (err error) {
		prf := []byte{prefixes.Event.B()}
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(prf); it.Next() {
			item := it.Item()
			if it.Item().ValueSize() == sha256.Size {
				continue
			}
			evKeys = append(evKeys, item.KeyCopy(nil))
		}
		return
	})
	var i int
	var key []byte
	for i, key = range evKeys {
		err = r.Update(func(txn *badger.Txn) (err error) {
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			it.Seek(key)
			if it.Valid() {
				item := it.Item()
				var evB []byte
				if evB, err = item.ValueCopy(nil); chk.E(err) {
					return
				}
				ser := serial.FromKey(key)
				var rem []byte
				ev := &event.T{}
				if rem, err = r.Unmarshal(ev, evB); chk.E(err) {
					return
				}
				if len(rem) > 0 {
					log.T.S(rem)
				}
				// 	add the indexes
				var indexKeys [][]byte
				indexKeys = GetIndexKeysForEvent(ev, ser)
				// log.I.S(indexKeys)
				for _, k := range indexKeys {
					var val []byte
					if k[0] == prefixes.Counter.B() {
						val = keys.Write(createdat.New(timestamp.Now()))
					}
					if err = txn.Set(k, val); chk.E(err) {
						return
					}
				}
				if i%1000 == 0 {
					log.I.F("rescanned %d events", i)
				}
			}
			return
		})
	}
	chk.E(err)
	log.I.F("completed rescanning %d events", i)
	return err
}
