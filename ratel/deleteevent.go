package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/keys/tombstone"
	"realy.lol/ratel/prefixes"
	"realy.lol/timestamp"
)

func (r *T) DeleteEvent(c context.T, eid *eventid.T, noTombstone ...bool) (err error) {
	var foundSerial []byte
	seri := serial.New(nil)
	err = r.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := prefixes.Id.Key(id.New(eid))
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(prf)
		if it.ValidForPrefix(prf) {
			var k []byte
			// get the serial
			k = it.Item().Key()
			// copy serial out
			keys.Read(k, index.Empty(), id.New(&eventid.T{}), seri)
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
	var indexKeys [][]byte
	ev := event.New()
	var evKey, evb, counterKey, tombstoneKey []byte
	// fetch the event to get its index keys
	err = r.View(func(txn *badger.Txn) (err error) {
		// retrieve the event record
		evKey = keys.Write(index.New(prefixes.Event), seri)
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(evKey)
		if it.ValidForPrefix(evKey) {
			if evb, err = it.Item().ValueCopy(evb); chk.E(err) {
				return
			}
			// log.I.S(evb)
			var rem []byte
			if rem, err = r.Unmarshal(ev, evb); chk.E(err) {
				return
			}
			if len(rem) != 0 {
				log.I.S(rem)
			}
			// log.I.S(rem, ev, seri)
			indexKeys = GetIndexKeysForEvent(ev, seri)
			counterKey = GetCounterKey(seri)
			// we don't make tombstones for replacements, but it is better to shift that
			// logic outside of this method.
			if len(noTombstone) > 0 && !noTombstone[0] {
				ts := tombstone.NewWith(ev.EventID())
				tombstoneKey = prefixes.Tombstone.Key(ts, createdat.New(timestamp.Now()))
			}
			return
		}
		return
	})
	if chk.E(err) {
		return
	}
	err = r.Update(func(txn *badger.Txn) (err error) {
		if err = txn.Delete(evKey); chk.E(err) {
		}
		for _, key := range indexKeys {
			if err = txn.Delete(key); chk.E(err) {
			}
		}
		if err = txn.Delete(counterKey); chk.E(err) {
			return
		}
		if len(tombstoneKey) > 0 {
			// write tombstone
			log.W.F("writing tombstone %0x for event %0x", tombstoneKey, ev.ID)
			if err = txn.Set(tombstoneKey, nil); chk.E(err) {
				return
			}
		}
		return
	})
	return
}
