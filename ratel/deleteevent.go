package ratel

import (
	"github.com/dgraph-io/badger/v4"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/keys/tombstone"
	"realy.lol/timestamp"
)

func (r *T) DeleteEvent(c Ctx, eid *eventid.T) (err E) {
	var foundSerial []byte
	seri := serial.New(nil)
	err = r.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := index.Id.Key(id.New(eid))
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(prf)
		if it.ValidForPrefix(prf) {
			var k []byte
			// get the serial
			k = it.Item().Key()
			// copy serial out
			keys.Read(k, index.Empty(), id.New(eventid.New()), seri)
			// save into foundSerial
			foundSerial = seri.Val
		}
		return
	})
	if chk.E(err) {
		return
	}
	if foundSerial == nil {
		return
	}
	var indexKeys []B
	ev := event.New()
	var evKey, evb, counterKey, tombstoneKey B
	// fetch the event to get its index keys
	err = r.View(func(txn *badger.Txn) (err error) {
		// retrieve the event record
		evKey = keys.Write(index.New(index.Event), seri)
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(evKey)
		if it.ValidForPrefix(evKey) {
			if evb, err = it.Item().ValueCopy(evb); chk.E(err) {
				return
			}
			// log.I.S(evb)
			var rem B
			if rem, err = ev.UnmarshalBinary(evb); chk.E(err) {
				return
			}
			_ = rem
			// log.I.S(rem, ev, seri)
			indexKeys = GetIndexKeysForEvent(ev, seri)
			counterKey = GetCounterKey(seri)
			ts := tombstone.NewWith(ev.EventID())
			tombstoneKey = index.Tombstone.Key(ts, createdat.New(timestamp.Now()))
			return
		}
		return
	})
	if chk.E(err) {
		return
	}
	err = r.Update(func(txn *badger.Txn) (err E) {
		if err = txn.Delete(evKey); chk.E(err) {
		}
		for _, key := range indexKeys {
			if err = txn.Delete(key); chk.E(err) {
			}
		}
		if err = txn.Delete(counterKey); chk.E(err) {
			return
		}
		// write tombstone
		if err = txn.Set(tombstoneKey, nil); chk.E(err) {
			return
		}
		return
	})
	return
}
