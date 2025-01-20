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
	"realy.lol/ratel/prefixes"
)

func (r *T) SaveEvent(c cx, ev *event.T) (err er) {
	if ev.Kind.IsEphemeral() {
		// log.T.F("not saving ephemeral event\n%s", ev.Serialize())
		return
	}
	// make sure Close waits for this to complete
	r.WG.Add(1)
	defer r.WG.Done()

	// first, search to see if the event ID already exists.
	var foundSerial by
	seri := serial.New(nil)
	var ts by
	txn := r.DB.NewTransaction(false)
	// query event by id to ensure we don't try to save duplicates
	prf := prefixes.Id.Key(id.New(eventid.NewWith(ev.ID)))
	it := txn.NewIterator(badger.IteratorOptions{})
	it.Seek(prf)
	if it.ValidForPrefix(prf) {
		var k by
		// get the serial
		k = it.Item().Key()
		// copy serial out
		keys.Read(k, index.Empty(), id.New(&eventid.T{}), seri)
		// save into foundSerial
		foundSerial = seri.Val
	}
	// if the event was deleted we don't want to save it again
	ts = prefixes.Tombstone.Key(id.New(eventid.NewWith(ev.ID)))
	it.Seek(ts)
	if it.ValidForPrefix(ts) {
		it.Close()
		txn.Discard()
		return errorf.W("tombstone found %0x, event will not be saved", ts)
	}
	it.Close()
	if err = txn.Commit(); chk.E(err) {
		return
	}

	// if we found something, check if it's a stub
	if foundSerial != nil {
		txn = r.DB.NewTransaction(true)
		// retrieve the event record
		evKey := keys.Write(index.New(prefixes.Event), seri)
		it = txn.NewIterator(badger.IteratorOptions{})
		bail := func() {
			it.Close()
			txn.Discard()
		}
		it.Seek(evKey)
		if it.ValidForPrefix(evKey) {
			if it.Item().ValueSize() != sha256.Size {
				// not a stub, we already have it
				bail()
				return eventstore.ErrDupEvent
			}
			// we only need to restore the event and write the access counter key
			var bin by
			bin = r.Marshal(ev, bin)
			if err = txn.Set(it.Item().Key(), bin); chk.E(err) {
				bail()
				return
			}
			// bump counter key
			counterKey := GetCounterKey(seri)
			val := keys.Write(createdat.New(timestamp.Now()))
			if err = txn.Set(counterKey, val); chk.E(err) {
				bail()
				return
			}
		}
		it.Close()
		if err = txn.Commit(); chk.E(err) {
			return
		}
		return
	}

	// now we can store the event
	var bin by
	bin = r.Marshal(ev, bin)
	// otherwise, save new event record.
	txn = r.DB.NewTransaction(true)
	var idx by
	var ser *serial.T
	idx, ser = r.SerialKey()
	// encode to binary
	// raw event store
	if err = txn.Set(idx, bin); chk.E(err) {
		txn.Discard()
		return
	}
	// 	add the indexes
	var indexKeys []by
	indexKeys = GetIndexKeysForEvent(ev, ser)
	// log.I.S(indexKeys)
	for _, k := range indexKeys {
		if err = txn.Set(k, nil); chk.E(err) {
			txn.Discard()
			return
		}
	}
	// initialise access counter key
	counterKey := GetCounterKey(ser)
	// log.I.S(counterKey)
	val := keys.Write(createdat.New(timestamp.Now()))
	if err = txn.Set(counterKey, val); chk.E(err) {
		txn.Discard()
		return
	}
	err = txn.Commit()
	return
}

func (r *T) Sync() (err er) { return r.DB.Sync() }
