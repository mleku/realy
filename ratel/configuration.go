package ratel

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v4"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/realy/config"
)

// SetConfiguration stores the store.C value to a provided setting.
func (r *T) SetConfiguration(c config.C) (err error) {
	var b []byte
	if b, err = json.Marshal(&c); chk.E(err) {
		return
	}
	err = r.Update(func(txn *badger.Txn) (err error) {
		if err = txn.Set(prefixes.Configuration.Key(), b); chk.E(err) {
			return
		}
		return
	})
	return
}

// GetConfiguration returns the current store.C stored in the database.
func (r *T) GetConfiguration() (c config.C, err error) {
	err = r.View(func(txn *badger.Txn) (err error) {
		c = config.C{BlockList: make([]string, 0)}
		var it *badger.Item
		if it, err = txn.Get(prefixes.Configuration.Key()); chk.E(err) {
			return
		}
		var b []byte
		if b, err = it.ValueCopy(nil); chk.E(err) {
			return
		}
		if err = json.Unmarshal(b, &c); chk.E(err) {
			return
		}
		return
	})
	return
}
