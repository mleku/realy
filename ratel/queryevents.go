package ratel

import (
	"github.com/dgraph-io/badger/v4"
	"mleku.dev/event"
	"mleku.dev/filter"
	"mleku.dev/ratel/keys/createdat"
	"mleku.dev/ratel/keys/index"
	"mleku.dev/ratel/keys/serial"
	"mleku.dev/sha256"
	"mleku.dev/tag"
)

func (r *T) QueryEvents(c Ctx, f *filter.T) (evs []*event.T, err E) {
	log.T.F("query for events\n%s", f)
	var queries []query
	var extraFilter *filter.T
	var since uint64
	if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
		return
	}
	var limit bool
	if f.Limit != 0 {
		log.T.S("query has a limit")
		limit = true
	}
	log.T.S(queries, extraFilter)
	// search for the keys generated from the filter
	var eventKeys [][]byte
	for _, q := range queries {
		log.T.S(q, extraFilter)
		err = r.View(func(txn *badger.Txn) (err E) {
			// iterate only through keys and in reverse order
			opts := badger.IteratorOptions{
				Reverse: true,
			}
			it := txn.NewIterator(opts)
			defer it.Close()
			// for it.Rewind(); it.Valid(); it.Next() {
			for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				log.T.S(k)
				if !q.skipTS {
					if len(k) < createdat.Len+serial.Len {
						continue
					}
					createdAt := createdat.FromKey(k)
					log.T.F("%d < %d", createdAt.Val.U64(), since)
					if createdAt.Val.U64() < since {
						break
					}
				}
				ser := serial.FromKey(k)
				eventKeys = append(eventKeys, index.Event.Key(ser))
			}
			return
		})
		if chk.E(err) {
			// this can't actually happen because the View function above does not set err.
		}
	search:
		for _, eventKey := range eventKeys {
			// log.I.S(eventKey)
			var v B
			err = r.View(func(txn *badger.Txn) (err E) {
				opts := badger.IteratorOptions{Reverse: true}
				it := txn.NewIterator(opts)
				defer it.Close()
				// for it.Rewind(); it.Valid(); it.Next() {
				for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
					item := it.Item()
					k := item.KeyCopy(nil)
					log.T.S(k)
					if v, err = item.ValueCopy(nil); chk.E(err) {
						continue
					}
					if r.HasL2 && len(v) == sha256.Size {
						// this is a stub entry that indicates an L2 needs to be accessed for it, so
						// we populate only the event.T.ID and return the result, the caller will
						// expect this as a signal to query the L2 event store.
						ev := &event.T{}
						log.T.F("found event stub %0x must seek in L2", v)
						ev.ID = v
						select {
						case <-c.Done():
							return
						case <-r.Ctx.Done():
							log.T.Ln("backend context canceled")
							return
						default:
						}
						evs = append(evs, ev)
						return
					}
				}
				return
			})
			if v == nil {
				continue
			}
			ev := &event.T{}
			var rem B
			if rem, err = ev.UnmarshalBinary(v); chk.E(err) {
				return
			}
			log.T.S(ev)
			if len(rem) > 0 {
				log.T.S(rem)
			}
			// check if this matches the other filters that were not part of the index.
			if extraFilter == nil || extraFilter.Matches(ev) {
				// check if this event is replaced by one we already have in the result.
				if ev.Kind.IsReplaceable() {
					for _, evc := range evs {
						// replaceable means there should be only the newest for the pubkey and
						// kind.
						if equals(ev.PubKey, evc.PubKey) && ev.Kind.Equal(evc.Kind) {
							// we won't add it to the results slice
							continue search
						}
					}
				}
				if ev.Kind.IsParameterizedReplaceable() &&
					ev.Tags.GetFirst(tag.New("d")) != nil {
					for _, evc := range evs {
						// parameterized replaceable means there should only be the newest for a
						// pubkey, kind and the value field of the `d` tag.
						if ev.Kind.Equal(evc.Kind) && equals(ev.PubKey, evc.PubKey) &&
							equals(ev.Tags.GetFirst(tag.New("d")).Value(),
								ev.Tags.GetFirst(tag.New("d")).Value()) {
							// we won't add it to the results slice
							continue search
						}
					}
				}
				log.T.F("sending back result\n%s\n", ev)
				evs = append(evs, ev)
				if limit {
					f.Limit--
					if f.Limit == 0 {
						return
					}
				} else {
					// if there is no limit, cap it at the MaxLimit, assume this was the intent
					// or the client is erroneous, if any limit greater is requested this will
					// be used instead as the previous clause.
					if len(evs) > r.MaxLimit {
						return
					}
				}
			}
		}
	}
	log.T.S(evs)
	log.T.Ln("query complete")
	return
}
