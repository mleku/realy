package ratel

import (
	"errors"
	"fmt"
	"io"

	"github.com/dgraph-io/badger/v4"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/qu"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/tags"
)

func (r *T) Export(c context.T, w io.Writer, pubkeys ...B) {
	var counter int
	var err error
	if len(pubkeys) > 0 {
		var pks []S
		for i := range pubkeys {
			pks = append(pks, hex.Enc(pubkeys[i]))
		}
		log.I.F("exporting selected pubkeys:\n%s", fmt.Sprint(pks))
		keyChan := make(chan B, 256)
		// specific set of public keys, so we need to run a search
		fa := &filter.T{Authors: tag.New(pubkeys...)}
		var queries []query
		if queries, _, _, err = PrepareQueries(fa); chk.E(err) {
			return
		}
		pTag := []B{B("#b")}
		pTag = append(pTag, pubkeys...)
		fp := &filter.T{Tags: tags.New(tag.New(pTag...))}
		var queries2 []query
		if queries2, _, _, err = PrepareQueries(fp); chk.E(err) {
			return
		}
		queries = append(queries, queries2...)
		log.I.S(queries)
		// start up writer loop
		quit := qu.T()
		go func() {
			for {
				select {
				case <-r.Ctx.Done():
					return
				case <-c.Done():
					return
				case <-quit:
					return
				case eventKey := <-keyChan:
					err = r.View(func(txn *badger.Txn) (err E) {
						select {
						case <-r.Ctx.Done():
							return
						case <-c.Done():
							return
						case <-quit:
							return
						default:
						}
						opts := badger.IteratorOptions{Reverse: true}
						it := txn.NewIterator(opts)
						defer it.Close()
						var count int
						for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
							count++
							item := it.Item()
							if r.HasL2 && item.ValueSize() == sha256.Size {
								// we aren't fetching from L2 for export, so don't send this
								// back.
								return
							}
							if err = item.Value(func(eventValue []byte) (err E) {
								ev := &event.T{}
								var rem B
								if rem, err = ev.UnmarshalBinary(eventValue); chk.E(err) {
									ev = nil
									eventValue = eventValue[:0]
									return
								}
								if len(rem) > 0 {
									log.T.S(rem)
								}
								if ev == nil {
									return
								}
								// log.I.Ln("found match", count, " for", pks)
								// send the event to client
								if _, err = fmt.Fprintf(w, "%s\n", ev.Serialize()); chk.E(err) {
									return
								}
								return
							}); chk.E(err) {
								return
							}
						}
						return
					})
					if chk.E(err) {
						// return
					}
				}
			}
		}()
		// stop the writer loop

		defer quit.Q()
		// log.I.Ln(len(queries), "queries for", pks)
		for _, q := range queries {
			select {
			case <-r.Ctx.Done():
				return
			case <-c.Done():
				return
			default:
			}
			// search for the keys generated from the filter
			err = r.View(func(txn *badger.Txn) (err E) {
				select {
				case <-r.Ctx.Done():
					return
				case <-c.Done():
					return
				default:
				}
				opts := badger.IteratorOptions{
					Reverse: true,
				}
				it := txn.NewIterator(opts)
				defer it.Close()
				for it.Seek(q.start); it.ValidForPrefix(q.searchPrefix); it.Next() {
					item := it.Item()
					k := item.KeyCopy(nil)
					evKey := index.Event.Key(serial.FromKey(k))
					counter++
					keyChan <- evKey
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
	} else {
		// blanket download requested
		err = r.View(func(txn *badger.Txn) (err error) {
			it := txn.NewIterator(badger.IteratorOptions{Prefix: index.Event.Key()})
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				select {
				case <-r.Ctx.Done():
					return
				case <-c.Done():
					return
				default:
				}
				item := it.Item()
				b, e := item.ValueCopy(nil)
				if e != nil {
					err = nil
					continue
				}
				var rem B
				ev := &event.T{}
				if rem, err = ev.UnmarshalBinary(b); chk.E(err) {
					err = nil
					continue
				}
				if len(rem) > 0 {
					log.T.S(rem)
				}
				// send the event to client
				if _, err = fmt.Fprintf(w, "%s\n", ev.Serialize()); chk.E(err) {
					return
				}
				counter++
			}
			return
		})
		chk.E(err)
	}
	log.I.Ln("exported", counter, "events")
	return
}
