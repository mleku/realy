package ratel

import (
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/log"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/realy/pointers"
	"realy.lol/tag"
	"realy.lol/timestamp"
)

func (r *T) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err error) {
	evMap := make(map[string]*event.T)
	var queries []query
	var ext *filter.T
	var since uint64
	if queries, ext, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	// log.I.S(f, queries)
	limit := r.MaxLimit
	if f.Limit != nil {
		limit = int(*f.Limit)
	}
	// search for the keys generated from the filter
	var total int
	eventKeys := make(map[string]struct{})
	for _, q := range queries {
		select {
		case <-r.Ctx.Done():
			return
		case <-c.Done():
			return
		default:
		}
		err = r.View(func(txn *badger.Txn) (err error) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				select {
				case <-r.Ctx.Done():
					return
				case <-c.Done():
					return
				default:
				}
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
				idx := prefixes.Event.Key(ser)
				eventKeys[string(idx)] = struct{}{}
				total++
				// some queries just produce stupid amounts of matches, they are a resource
				// exhaustion attack vector and only spiders make them
				if total >= limit {
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
	select {
	case <-r.Ctx.Done():
		return
	case <-c.Done():
		return
	default:
	}
	var delEvs [][]byte
	defer func() {
		for _, d := range delEvs {
			// if events were found that should be deleted, delete them
			chk.E(r.DeleteEvent(r.Ctx, eventid.NewWith(d)))
		}
	}()
	accessed := make(map[string]struct{})
	for ek := range eventKeys {
		eventKey := []byte(ek)
		err = r.View(func(txn *badger.Txn) (err error) {
			select {
			case <-r.Ctx.Done():
				return
			case <-c.Done():
				return
			default:
			}
			opts := badger.IteratorOptions{Reverse: true}
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
				item := it.Item()
				var ev *event.T
				if delEvs, ev, err = r.ProcessFoundEvent(item, delEvs); chk.E(err) {
					continue
				}
				if ev == nil {
					continue
				}
				if ext == nil || ext.Matches(ev) {
					evMap[hex.Enc(ev.Id)] = ev
					// add event counter key to accessed
					ser := serial.FromKey(eventKey)
					accessed[string(ser.Val)] = struct{}{}
					if pointers.Present(f.Limit) {
						*f.Limit--
						if *f.Limit <= 0 {
							log.I.F("found events: %d", len(evMap))
							return
						}
					}
					// if there is no limit, cap it at the MaxLimit, assume this was the
					// intent or the client is erroneous, if any limit greater is
					// requested this will be used instead as the previous clause.
					if len(evMap) >= r.MaxLimit {
						return
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
		select {
		case <-r.Ctx.Done():
			return
		case <-c.Done():
			return
		default:
		}
	}
	if len(evMap) > 0 {
		evs = r.FilterSortAndLimit(evMap, limit)
		go r.UpdateAccessed(accessed)
	} else {
		log.T.F("no events found,%s", f.Serialize())
	}
	return
}

func (r *T) FilterSortAndLimit(evMap map[string]*event.T, limit int) (evs event.Ts) {
	for i := range evMap {
		if len(evMap[i].Pubkey) == 0 {
			log.I.S(evMap[i])
			continue
		}
		evs = append(evs, evMap[i])
	}
	sort.Sort(event.Descending(evs))
	if len(evs) > limit {
		evs = evs[:limit]
	}
	return
}

func (r *T) ProcessFoundEvent(item *badger.Item, delEvs [][]byte) (dEvs [][]byte, ev *event.T, err error) {
	err = item.Value(func(eventValue []byte) (err error) {
		var rem []byte
		ev = &event.T{}
		if rem, err = ev.Unmarshal(eventValue); chk.E(err) {
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
				ev = nil
				return
			}
		}
		return
	})
	return
}

func (r *T) UpdateAccessed(accessed map[string]struct{}) {
	var err error
	now := timestamp.Now()
	for ser := range accessed {
		if err = r.Update(func(txn *badger.Txn) (err error) {
			seri := serial.New([]byte(ser))
			key := GetCounterKey(seri)
			it := txn.NewIterator(badger.IteratorOptions{})
			defer it.Close()
			if it.Seek(key); it.ValidForPrefix(key) {
				// update access record
				if err = txn.Set(key, now.Bytes()); chk.E(err) {
					return
				}
			}
			return nil
		}); chk.E(err) {
			return
		}
	}
}
