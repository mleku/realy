package ratel

import (
	"errors"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/ratel/prefixes"
)

func (r *T) CountEvents(c cx, f *filter.T) (count no, approx bo, err er) {
	log.T.F("QueryEvents,%s", f.Serialize())
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	// search for the keys generated from the filter
	for _, q := range queries {
		select {
		case <-c.Done():
			return
		default:
		}
		var eventKey by
		err = r.View(func(txn *badger.Txn) (err er) {
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
				// todo: here we should get the kind field from the key and and collate the
				// todo: matches that are replaceable/parameterized replaceable ones to decode
				// todo: to check for replacements so we can actually not set the approx flag.
				ser := serial.FromKey(k)
				eventKey = prefixes.Event.Key(ser)
				// eventKeys = append(eventKeys, idx)
			}
			return
		})
		if chk.E(err) {
			// this means shutdown, probably
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
		}
		// todo: here we should decode replaceable events and discard the outdated versions
		if extraFilter != nil {
			// if there is an extra filter we need to fetch and decode the event to determine a
			// match.
			err = r.View(func(txn *badger.Txn) (err er) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					if r.HasL2 && item.ValueSize() == sha256.Size {
						// we will count this though it may not match in fact. for general,
						// simple filters there isn't likely to be an extrafilter anyway. the
						// count result can have an "approximate" flag so we flip this now.
						approx = true
						return
					}
					ev := &event.T{}
					var appr bo
					if err = item.Value(func(eventValue by) (err er) {
						var rem by
						if rem, err = r.Unmarshal(ev, eventValue); chk.E(err) {
							return
						}
						if len(rem) > 0 {
							log.T.S(rem)
						}
						if ev.Kind.IsReplaceable() ||
							(ev.Kind.IsParameterizedReplaceable() &&
								ev.Tags.GetFirst(tag.New("d")) != nil) {
							// we aren't going to spend this extra time so this just flips the
							// approximate flag. generally clients are asking for counts to get
							// an outside estimate anyway, to avoid exceeding MaxLimit
							appr = true
						}
						return
					}); chk.E(err) {
						continue
					}
					if ev == nil {
						continue
					}
					if extraFilter.Matches(ev) {
						count++
						if appr {
							approx = true
						}
						return
					}
				}
				return
			})
		} else {
			count++
		}
	}
	return
}
