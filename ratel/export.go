package ratel

import (
	"errors"
	"fmt"
	"io"

	"github.com/dgraph-io/badger/v4"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/log"
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/sha256"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/tags"
)

// Export the complete database of stored events to an io.Writer in line structured minified
// JSON.
func (r *T) Export(c context.T, w io.Writer, pubkeys ...[]byte) {
	var counter int
	var err error
	if len(pubkeys) > 0 {
		var pks []string
		for i := range pubkeys {
			pks = append(pks, hex.Enc(pubkeys[i]))
		}
		o := "["
		for _, pk := range pks {
			o += pk + ","
		}
		o += "]"
		log.I.F("exporting selected pubkeys:\n%s", o)
		keyChan := make(chan []byte, 256)
		// specific set of public keys, so we need to run a search
		fa := &filter.T{Authors: tag.New(pubkeys...)}
		var queries []query
		if queries, _, _, err = PrepareQueries(fa); chk.E(err) {
			return
		}
		pTag := [][]byte{[]byte("#b")}
		pTag = append(pTag, pubkeys...)
		fp := &filter.T{Tags: tags.New(tag.New(pTag...))}
		var queries2 []query
		if queries2, _, _, err = PrepareQueries(fp); chk.E(err) {
			return
		}
		queries = append(queries, queries2...)
		// start up writer loop
		quit := make(chan struct{})
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
					err = r.View(func(txn *badger.Txn) (err error) {
						select {
						case <-r.Ctx.Done():
							return
						case <-c.Done():
							return
						case <-quit:
							return
						default:
						}
						opts := badger.IteratorOptions{Reverse: false}
						it := txn.NewIterator(opts)
						defer it.Close()
						var count int
						for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
							count++
							item := it.Item()
							if r.HasL2 && item.ValueSize() == sha256.Size {
								// we aren't fetching from L2 for export, so don't send this back.
								return
							}
							if err = item.Value(func(eventValue []byte) (err error) {
								// send the event to client (no need to re-encode it)
								if _, err = fmt.Fprintf(w, "%s\n", eventValue); chk.E(err) {
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
					}
				}
			}
		}()
		// stop the writer loop
		defer close(quit)
		for _, q := range queries {
			select {
			case <-r.Ctx.Done():
				return
			case <-c.Done():
				return
			default:
			}
			// search for the keys generated from the filter
			err = r.View(func(txn *badger.Txn) (err error) {
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
					evKey := prefixes.Event.Key(serial.FromKey(k))
					counter++
					if counter%1000 == 0 && counter > 0 {
						log.I.F("%d events exported", counter)
					}
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
			it := txn.NewIterator(badger.IteratorOptions{Prefix: prefixes.Event.Key()})
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
				if chk.E(e) {
					// already isn't the same as the return value!
					// err = nil
					continue
				}
				// send the event to client
				if r.UseCompact {
					ev := &event.T{}
					var rem []byte
					rem, err = ev.UnmarshalCompact(b)
					if chk.E(err) {
						err = nil
						continue
					}
					if len(rem) > 0 {
						log.I.S(rem)
					}
					if _, err = fmt.Fprintf(w, "%s\n", ev.Marshal(nil)); chk.E(err) {
						return
					}

				} else {
					// the database stores correct JSON versions so no need to decode/encode.
					if _, err = fmt.Fprintf(w, "%s\n", b); chk.E(err) {
						return
					}
				}
				counter++
				if counter%1000 == 0 && counter > 0 {
					log.I.F("%d events exported", counter)
				}
			}
			return
		})
		chk.E(err)
	}
	log.I.Ln("exported", counter, "events")
	return
}
