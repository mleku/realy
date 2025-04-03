package ratel

import (
	"errors"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
	"realy.lol/ratel/keys/fullpubkey"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/realy/pointers"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/timestamp"
)

func (r *T) QueryForIds(c context.T, f *filter.T) (founds []store.IdTsPk, err error) {
	log.T.F("QueryForIds %s\n", f.Serialize())
	var queries []query
	var ext *filter.T
	var since uint64
	if queries, ext, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	// search for the keys generated from the filter
	var total int
	eventKeys := make(map[string]struct{})
	var serials []*serial.T
	for _, q := range queries {
		err = r.View(func(txn *badger.Txn) (err error) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				if !q.skipTS {
					if len(k) < createdat.Len+serial.Len {
						continue
					}
					createdAt := createdat.FromKey(k)
					if createdAt.Val.U64() < since {
						break
					}
				}
				ser := serial.FromKey(k)
				serials = append(serials, ser)
				idx := prefixes.Event.Key(ser)
				eventKeys[string(idx)] = struct{}{}
				total++
				// some queries just produce stupid amounts of matches, they are a resource
				// exhaustion attack vector and only spiders make them
				if total > 5000 {
					return
				}
			}
			return
		})
		if chk.E(err) {
			// this means shutdown, probably
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
		}
	}
	log.T.F("found %d event indexes from %d queries", len(eventKeys), len(queries))
	l2Map := make(map[string]*event.T) // todo: this is not being used, it should be
	var delEvs [][]byte
	defer func() {
		for _, d := range delEvs {
			// if events were found that should be deleted, delete them
			chk.E(r.DeleteEvent(r.Ctx, eventid.NewWith(d)))
		}
	}()
	accessed := make(map[string]struct{})
	if ext != nil {
		// we have to fetch the event
		for ek := range eventKeys {
			eventKey := []byte(ek)
			err = r.View(func(txn *badger.Txn) (err error) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
			done:
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					if r.HasL2 && item.ValueSize() == sha256.Size {
						// this is a stub entry that indicates an L2 needs to be accessed for
						// it, so we populate only the event.T.Id and return the result, the
						// caller will expect this as a signal to query the L2 event store.
						var eventValue []byte
						ev := &event.T{}
						if eventValue, err = item.ValueCopy(nil); chk.E(err) {
							continue
						}
						log.T.F("found event stub %0x must seek in L2", eventValue)
						ev.Id = eventValue
						l2Map[hex.Enc(ev.Id)] = ev
						return
					}
					ev := &event.T{}
					if err = item.Value(func(eventValue []byte) (err error) {
						var rem []byte
						if rem, err = r.Unmarshal(ev, eventValue); chk.E(err) {
							return
						}
						if len(rem) > 0 {
							log.T.S(rem)
						}
						if et := ev.Tags.GetFirst(tag.New("expiration")); et != nil {
							var exp uint64
							if exp, err = strconv.ParseUint(string(et.Value()), 10,
								64); chk.E(err) {
								return
							}
							if int64(exp) > time.Now().Unix() {
								// this needs to be deleted
								delEvs = append(delEvs, ev.Id)
								return
							}
						}
						return
					}); chk.E(err) {
						continue
					}
					if ev == nil {
						continue
					}
					if ext.Matches(ev) {
						// add event counter key to accessed
						ser := serial.FromKey(eventKey)
						serials = append(serials, ser)
						accessed[string(ser.Val)] = struct{}{}
						if pointers.Present(f.Limit) {
							if *f.Limit < uint(len(serials)) {
								// done
								break done
							}
						}
					}
				}
				return
			})
			if err != nil {
				// this means shutdown, probably
				if errors.Is(err, badger.ErrDBClosed) {
					return
				}
			}
		}
	}
	for _, ser := range serials {
		err = r.View(func(txn *badger.Txn) (err error) {
			prf := prefixes.FullIndex.Key(ser)
			opts := badger.IteratorOptions{Prefix: prf}
			it := txn.NewIterator(opts)
			defer it.Close()
			it.Seek(prf)
			if it.ValidForPrefix(prf) {
				k := it.Item().KeyCopy(nil)
				id := fullid.New()
				ts := createdat.New(timestamp.New())
				pk := fullpubkey.New()
				keys.Read(k, index.New(0), serial.New(nil), id, pk, ts)
				ff := store.IdTsPk{
					Ts:  ts.Val.I64(),
					Id:  id.Val,
					Pub: pk.Val,
				}
				founds = append(founds, ff)
			}
			return
		})
	}
	// log.I.S(founds)
	return
}
