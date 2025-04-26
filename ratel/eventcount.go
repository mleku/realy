package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/ratel/prefixes"
)

func (r *T) EventCount() (count uint64, err error) {
	err = r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: prefixes.Event.Key()})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return
	})
	return
}
