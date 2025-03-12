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
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/sha256"
	"realy.lol/tag"
)

func (r *T) CountEvents(c context.T, f *filter.T) (count int, approx bool, err error) {
	log.T.F("QueryEvents,%s", f.Serialize())
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	var delEvs [][]byte
	defer func() {
		// after the count delete any events that are expired as per NIP-40
		for _, d := range delEvs {
			chk.E(r.DeleteEvent(r.Ctx, eventid.NewWith(d)))
		}
	}()
	// search for the keys generated from the filter
	for _, q := range queries {
		select {
		case <-c.Done():
			return
		default:
		}
		var eventKey []byte
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
			err = r.View(func(txn *badger.Txn) (err error) {
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
					var appr bool
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
