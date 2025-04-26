package ratel

import (
	"io"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/tag"
)

// FetchIds retrieves events based on a list of event Ids that have been provided.
func (r *T) FetchIds(c context.T, evIds *tag.T, out io.Writer) (err error) {
	b := make([]byte, 0, 100000)
	err = r.View(func(txn *badger.Txn) (err error) {
		for _, v := range evIds.ToSliceOfBytes() {
			var evId *id.T
			if evId, err = id.NewFromBytes(v); chk.E(err) {
				return
			}
			k := prefixes.Id.Key(evId)
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			var ser *serial.T
			for it.Seek(k); it.ValidForPrefix(k); it.Next() {
				key := it.Item().Key()
				ser = serial.FromKey(key)
				break
			}
			var item *badger.Item
			if item, err = txn.Get(prefixes.Event.Key(ser)); chk.E(err) {
				return
			}
			if b, err = item.ValueCopy(nil); chk.E(err) {
				return
			}
			if _, err = out.Write(b); chk.E(err) {
				return
			}
			// add the new line after entries
			if _, err = out.Write([]byte{'\n'}); chk.E(err) {
				return
			}
		}
		return
	})
	return
}
