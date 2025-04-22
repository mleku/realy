package ratel

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v4"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/log"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/store"
)

// SetConfiguration stores the store.Configuration value to a provided setting.
func (r *T) SetConfiguration(c *store.Configuration) (err error) {
	var b []byte
	if b, err = json.Marshal(c); chk.E(err) {
		return
	}
	log.I.F("%s", b)
	err = r.Update(func(txn *badger.Txn) (err error) {
		if err = txn.Set(prefixes.Configuration.Key(), b); chk.E(err) {
			return
		}
		return
	})
	return
}

// GetConfiguration returns the current store.Configuration stored in the database.
func (r *T) GetConfiguration() (c *store.Configuration, err error) {
	err = r.View(func(txn *badger.Txn) (err error) {
		c = &store.Configuration{BlockList: make([]string, 0)}
		var it *badger.Item
		if it, err = txn.Get(prefixes.Configuration.Key()); chk.E(err) {
			err = nil
			return
		}
		var b []byte
		if b, err = it.ValueCopy(nil); chk.E(err) {
			return
		}
		if err = json.Unmarshal(b, c); chk.E(err) {
			return
		}
		return
	})
	return
}
