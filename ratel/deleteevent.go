package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/log"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/keys/tombstone"
	"realy.lol/ratel/prefixes"
	"realy.lol/timestamp"
)

// DeleteEvent deletes an event if it exists, and writes a tombstone for the event unless
// requested not to, so that the event can't be saved again.
func (r *T) DeleteEvent(c context.T, eid *eventid.T, noTombstone ...bool) (err error) {
	var foundSerial []byte
	ser := serial.New(nil)
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
			keys.Read(k, index.Empty(), id.New(&eventid.T{}), ser)
			// save into foundSerial
			foundSerial = ser.Val
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
	var evKey, evb, tombstoneKey []byte
	var w, l [][]byte
	// fetch the event to get its index keys
	err = r.View(func(txn *badger.Txn) (err error) {
		// retrieve the event record
		evKey = keys.Write(index.New(prefixes.Event), ser)
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		it.Seek(evKey)
		if it.ValidForPrefix(evKey) {
			if evb, err = it.Item().ValueCopy(evb); chk.E(err) {
				return
			}
			if _, err = r.Unmarshal(ev, evb); chk.E(err) {
				return
			}
			indexKeys = GetIndexKeysForEvent(ev, ser)
			indexKeys = append(indexKeys, r.GetFulltextKeys(ev, ser)...)
			indexKeys = append(indexKeys, r.GetLangKeys(ev, ser)...)
			// we don't make tombstones for replacements, but it is better to shift that
			// logic outside of this closure.
			if len(noTombstone) > 0 && !noTombstone[0] {
				ts := tombstone.NewWith(ev.EventId())
				tombstoneKey = prefixes.Tombstone.Key(ts, createdat.New(timestamp.Now()))
			}
			return
		}
		return
	})
	if chk.E(err) {
		return
	}
	_, _ = w, l

	err = r.Update(func(txn *badger.Txn) (err error) {
		if err = txn.Delete(evKey); chk.E(err) {
		}
		for _, key := range indexKeys {
			if err = txn.Delete(key); chk.E(err) {
			}
		}
		if len(tombstoneKey) > 0 {
			// write tombstone
			log.W.F("writing tombstone %0x for event %0x", tombstoneKey, ev.Id)
			if err = txn.Set(tombstoneKey, nil); chk.E(err) {
				return
			}
		}
		return
	})
	return
}
