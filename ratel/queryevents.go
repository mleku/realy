package ratel

import (
	"errors"
	"fmt"
	"sort"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/timestamp"
)

func (r *T) QueryEvents(c cx, f *filter.T) (evs event.Ts, err er) {
	log.T.F("QueryEvents,%s", f.Serialize())
	evMap := make(map[st]*event.T)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	// search for the keys generated from the filter
	eventKeys := make(map[st]struct{})
	for _, q := range queries {
		select {
		case <-r.Ctx.Done():
			return
		case <-c.Done():
			return
		default:
		}
		err = r.View(func(txn *badger.Txn) (err er) {
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
				idx := index.Event.Key(ser)
				eventKeys[st(idx)] = struct{}{}
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
	log.T.F("found %d event indexes", len(eventKeys))
	select {
	case <-r.Ctx.Done():
		return
	case <-c.Done():
		return
	default:
	}
	accessed := make(map[st]struct{})
	for ek := range eventKeys {
		eventKey := by(ek)
		var done bo
		err = r.View(func(txn *badger.Txn) (err er) {
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
				if r.HasL2 && item.ValueSize() == sha256.Size {
					// this is a stub entry that indicates an L2 needs to be accessed for
					// it, so we populate only the event.T.ID and return the result, the
					// caller will expect this as a signal to query the L2 event store.
					var eventValue by
					ev := &event.T{}
					if eventValue, err = item.ValueCopy(nil); chk.E(err) {
						continue
					}
					log.T.F("found event stub %0x must seek in L2", eventValue)
					ev.ID = eventValue
					select {
					case <-c.Done():
						return
					case <-r.Ctx.Done():
						log.T.Ln("backend context canceled")
						return
					default:
					}
					evMap[hex.Enc(ev.ID)] = ev
					return
				}
				ev := &event.T{}
				if err = item.Value(func(eventValue by) (err er) {
					var rem by
					if rem, err = ev.Unmarshal(eventValue); chk.E(err) {
						ev = nil
						eventValue = eventValue[:0]
						return
					}
					if len(rem) > 0 {
						log.T.S(rem)
					}
					// check if this event is replaced by one we already have in the result.
					if ev.Kind.IsReplaceable() {
						for i, evc := range evMap {
							// replaceable means there should be only the newest for the
							// pubkey and kind.
							if equals(ev.PubKey, evc.PubKey) && ev.Kind.Equal(evc.Kind) {
								if ev.CreatedAt.I64() > evc.CreatedAt.I64() {
									// log.T.F("event %0x,%s\nreplaces %0x,%s",
									// 	ev.ID, ev.Serialize(),
									// 	evc.ID, evc.Serialize(),
									// )
									// replace the event, it is newer
									delete(evMap, i)
									break
								} else {
									// we won't add it to the results slice
									eventValue = eventValue[:0]
									ev = nil
									return
								}
							}
						}
					} else if ev.Kind.IsParameterizedReplaceable() &&
						ev.Tags.GetFirst(tag.New("d")) != nil {
						for i, evc := range evMap {
							// parameterized replaceable means there should only be the
							// newest for a pubkey, kind and the value field of the `d` tag.
							if ev.Kind.Equal(evc.Kind) && equals(ev.PubKey, evc.PubKey) &&
								equals(ev.Tags.GetFirst(tag.New("d")).Value(),
									ev.Tags.GetFirst(tag.New("d")).Value()) {
								if ev.CreatedAt.I64() > evc.CreatedAt.I64() {
									log.T.F("event %0x,%s\nreplaces %0x,%s",
										ev.ID,
										ev.Serialize(),
										evc.ID,
										evc.Serialize(),
									)
									// replace the event, it is newer
									delete(evMap, i)
									break
								} else {
									// we won't add it to the results slice
									eventValue = eventValue[:0]
									ev = nil
									return
								}
							}
						}
					}
					return
				}); chk.E(err) {
					continue
				}
				if ev == nil {
					continue
				}
				if extraFilter == nil || extraFilter.Matches(ev) {
					evMap[hex.Enc(ev.ID)] = ev
					// add event counter key to accessed
					ser := serial.FromKey(eventKey)
					accessed[st(ser.Val)] = struct{}{}
					if filter.Present(f.Limit) {
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
						log.T.F("found MaxLimit events: %d", len(evMap))
						done = true
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
		if done {
			err = nil
			return
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
		for i := range evMap {
			if len(evMap[i].PubKey) == 0 {
				continue
			}
			evs = append(evs, evMap[i])
		}
		sort.Sort(event.Descending(evs))
		log.T.C(func() string {
			evIds := make([]string, len(evs))
			for i, ev := range evs {
				evIds[i] = hex.Enc(ev.ID)
			}
			heading := fmt.Sprintf("query complete,%d events found,%s", len(evs),
				f.Serialize())
			return fmt.Sprintf("%s\nevents,%v", heading, evIds)
		})
		// bump the access times on all retrieved events. do this in a goroutine so the
		// user's events are delivered immediately
		go func() {
			for ser := range accessed {
				seri := serial.New(by(ser))
				now := timestamp.Now()
				err = r.Update(func(txn *badger.Txn) (err er) {
					key := GetCounterKey(seri)
					it := txn.NewIterator(badger.IteratorOptions{})
					defer it.Close()
					if it.Seek(key); it.ValidForPrefix(key) {
						// update access record
						if err = txn.Set(key, now.Bytes()); chk.E(err) {
							return
						}
					}
					// log.T.Ln("last access for", seri.Uint64(), now.U64())
					return nil
				})
			}
		}()
	} else {
		log.T.F("no events found,%s", f.Serialize())
	}
	// }
	return
}
