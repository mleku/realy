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
	"realy.lol/sha256"
	eventstore "realy.lol/store"
	"realy.lol/timestamp"
)

func (r *T) SaveEvent(c Ctx, ev *event.T) (err E) {
	if ev.Kind.IsEphemeral() {
		// log.T.F("not saving ephemeral event\n%s", ev.Serialize())
		return
	}
	// make sure Close waits for this to complete
	r.WG.Add(1)
	defer r.WG.Done()
	// first, search to see if the event ID already exists.
	var foundSerial []byte
	var deleted bool
	seri := serial.New(nil)
	var ts B
	err = r.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := index.Id.Key(id.New(eventid.NewWith(ev.ID)))
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
		// if the event was deleted we don't want to save it again
		ts = index.Tombstone.Key(id.New(eventid.NewWith(ev.ID)))
		it.Seek(ts)
		if it.ValidForPrefix(ts) {
			deleted = true
		}
		return
	})
	if chk.E(err) {
		return
	}
	if deleted {
		return errorf.W("tombstone found %0x, event will not be saved", ts)
	}
	if foundSerial != nil {
		// log.D.F("found possible duplicate or stub for %s", ev.Serialize())
		err = r.Update(func(txn *badger.Txn) (err error) {
			// retrieve the event record
			evKey := keys.Write(index.New(index.Event), seri)
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			it.Seek(evKey)
			if it.ValidForPrefix(evKey) {
				if it.Item().ValueSize() != sha256.Size {
					// not a stub, we already have it
					// log.D.F("duplicate event %0x", ev.ID)
					return eventstore.ErrDupEvent
				}
				// we only need to restore the event binary and write the access counter key
				// encode to binary
				var bin B
				if bin, err = ev.MarshalBinary(bin); chk.E(err) {
					return
				}
				if err = txn.Set(it.Item().Key(), bin); chk.E(err) {
					return
				}
				// bump counter key
				counterKey := GetCounterKey(seri)
				val := keys.Write(createdat.New(timestamp.Now()))
				if err = txn.Set(counterKey, val); chk.E(err) {
					return
				}
				return
			}
			return
		})
		// if it was a dupe, we are done.
		if err != nil {
			return
		}
		return
	}
	var bin B
	if bin, err = ev.MarshalBinary(bin); chk.T(err) {
		return
	}
	// otherwise, save new event record.
	if err = r.Update(func(txn *badger.Txn) (err error) {
		var idx []byte
		var ser *serial.T
		idx, ser = r.SerialKey()
		// encode to binary
		// raw event store
		if err = txn.Set(idx, bin); chk.E(err) {
			return
		}
		// 	add the indexes
		var indexKeys [][]byte
		indexKeys = GetIndexKeysForEvent(ev, ser)
		// log.I.S(indexKeys)
		for _, k := range indexKeys {
			if err = txn.Set(k, nil); chk.E(err) {
				return
			}
		}
		// initialise access counter key
		counterKey := GetCounterKey(ser)
		// log.I.S(counterKey)
		val := keys.Write(createdat.New(timestamp.Now()))
		if err = txn.Set(counterKey, val); chk.E(err) {
			return
		}
		log.D.F("saved event to ratel %s:\n%s", r.dataDir, ev.Serialize())
		return
	}); chk.E(err) {
		return
	}
	return
}

func (r *T) Sync() (err E) { return r.DB.Sync() }
