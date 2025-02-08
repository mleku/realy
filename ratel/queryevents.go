package ratel

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/timestamp"
)

func (r *T) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err error) {
	log.T.F("QueryEvents %s", f.Serialize())
	evMap := make(map[string]*event.T)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
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
			chk.E(r.DeleteEvent(r.Ctx, eventid.NewWith(d)))
		}
	}()
	accessed := make(map[string]struct{})
	for ek := range eventKeys {
		eventKey := []byte(ek)
		var done bool
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
				if r.HasL2 && item.ValueSize() == sha256.Size {
					// this is a stub entry that indicates an L2 needs to be accessed for
					// it, so we populate only the event.T.ID and return the result, the
					// caller will expect this as a signal to query the L2 event store.
					var eventValue []byte
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
						if exp, err = strconv.ParseUint(string(et.Value()), 10, 64); chk.E(err) {
							return
						}
						if int64(exp) > time.Now().Unix() {
							// this needs to be deleted
							delEvs = append(delEvs, ev.ID)
							return
						}
					}
					// check if this event is replaced by one we already have in the result.
					if ev.Kind.IsReplaceable() {
						for i, evc := range evMap {
							// replaceable means there should be only the newest for the
							// pubkey and kind.
							if bytes.Equal(ev.PubKey, evc.PubKey) && ev.Kind.Equal(evc.Kind) {
								if ev.CreatedAt.I64() > evc.CreatedAt.I64() {
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
							if ev.Kind.Equal(evc.Kind) && bytes.Equal(ev.PubKey, evc.PubKey) &&
								bytes.Equal(ev.Tags.GetFirst(tag.New("d")).Value(),
									evc.Tags.GetFirst(tag.New("d")).Value()) {
								if ev.CreatedAt.I64() > evc.CreatedAt.I64() {
									log.T.F("event %0x,%s\n->replaces\n%0x,%s",
										ev.ID, ev.Serialize(),
										evc.ID, evc.Serialize(),
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
					accessed[string(ser.Val)] = struct{}{}
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
						// log.T.F("found MaxLimit events: %d", len(evMap))
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
				log.I.S(evMap[i])
				continue
			}
			evs = append(evs, evMap[i])
		}
		sort.Sort(event.Descending(evs))
		if len(evs) > limit {
			evs = evs[:limit]
		}
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
				seri := serial.New([]byte(ser))
				now := timestamp.Now()
				err = r.Update(func(txn *badger.Txn) (err error) {
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
