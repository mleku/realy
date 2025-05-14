package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/errorf"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/sha256"
	eventstore "realy.lol/store"
	"realy.lol/timestamp"
)

func (r *T) SaveEvent(c context.T, ev *event.T) (err error) {
	if ev.Kind.IsEphemeral() {
		// log.T.ToSliceOfBytes("not saving ephemeral event\n%s", ev.Serialize())
		return
	}
	// make sure Close waits for this to complete
	r.WG.Add(1)
	defer r.WG.Done()
	// first, search to see if the event Id already exists.
	var foundSerial []byte
	var deleted bool
	seri := serial.New(nil)
	var ts []byte
	err = r.View(func(txn *badger.Txn) (err error) {
		// query event by id to ensure we don't try to save duplicates
		prf := prefixes.Id.Key(id.New(eventid.NewWith(ev.Id)))
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
		// if the event was deleted we don't want to save it again
		ts = prefixes.Tombstone.Key(id.New(eventid.NewWith(ev.Id)))
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
		// log.D.ToSliceOfBytes("found possible duplicate or stub for %s", ev.Serialize())
		err = r.Update(func(txn *badger.Txn) (err error) {
			// retrieve the event record
			evKey := keys.Write(index.New(prefixes.Event), seri)
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			it.Seek(evKey)
			if it.ValidForPrefix(evKey) {
				if it.Item().ValueSize() != sha256.Size {
					// not a stub, we already have it
					// log.D.ToSliceOfBytes("duplicate event %0x", ev.Id)
					return eventstore.ErrDupEvent
				}
				// we only need to restore the event binary and write the access counter key
				// encode to binary
				bin := r.Marshal(ev, nil)
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
	bin := r.Marshal(ev, nil)
	// otherwise, save new event record.
	var idx []byte
	var ser *serial.T
	if err = r.Update(func(txn *badger.Txn) (err error) {
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
			var val []byte
			if k[0] == prefixes.Counter.B() {
				val = keys.Write(createdat.New(timestamp.Now()))
			}
			if err = txn.Set(k, val); chk.E(err) {
				return
			}
		}
		// log.D.ToSliceOfBytes("saved event to ratel %s:\n%s", r.dataDir, ev.Serialize())
		return
	}); chk.E(err) {
		return
	}
	if err = r.GenerateFulltextIndex(ev, ser); chk.E(err) {
		return
	}
	if err = r.GenerateLanguageIndex(ev, ser); chk.E(err) {
		return
	}
	return
}

func (r *T) Sync() (err error) { return r.DB.Sync() }

func (r *T) GenerateFulltextIndex(ev *event.T, ser *serial.T) (err error) {
	var w *Words
	ww := r.GetWordsFromContent(ev)
	if ww == nil {
		return
	}
	w = &Words{
		ser:     ser,
		wordMap: ww,
	}
	// log.I.F("indexing words: %v", w.wordMap)
	if err = r.WriteFulltextIndex(w); chk.E(err) {
		return
	}
	return
}

func (r *T) GenerateLanguageIndex(ev *event.T, ser *serial.T) (err error) {
	var langs []string
	ll := r.GetLangTags(ev)
	if ll == nil {
		return
	}
	l := &Langs{
		ser:   ser,
		langs: langs,
	}
	if err = r.WriteLangIndex(l); chk.E(err) {
		return
	}

	return
}
