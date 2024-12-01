package ratel

import (
	"errors"
	"fmt"
	"io"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/qu"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/tag"
	"realy.lol/tags"
)

func (r *T) Export(c cx, w io.Writer, pubkeys ...by) {
	var counter no
	var err er
	if len(pubkeys) > 0 {
		var pks []st
		for i := range pubkeys {
			pks = append(pks, hex.Enc(pubkeys[i]))
		}
		o := "["
		for _, pk := range pks {
			o += pk + ","
		}
		o += "]"
		log.I.F("exporting selected pubkeys:\n%s", o)
		keyChan := make(chan by, 256)
		// specific set of public keys, so we need to run a search
		fa := &filter.T{Authors: tag.New(pubkeys...)}
		var queries []query
		if queries, _, _, err = PrepareQueries(fa); chk.E(err) {
			return
		}
		pTag := []by{by("#b")}
		pTag = append(pTag, pubkeys...)
		fp := &filter.T{Tags: tags.New(tag.New(pTag...))}
		var queries2 []query
		if queries2, _, _, err = PrepareQueries(fp); chk.E(err) {
			return
		}
		queries = append(queries, queries2...)
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
					err = r.View(func(txn *badger.Txn) (err er) {
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
						var count no
						for it.Seek(eventKey); it.ValidForPrefix(eventKey); it.Next() {
							count++
							item := it.Item()
							if r.HasL2 && item.ValueSize() == sha256.Size {
								// we aren't fetching from L2 for export, so don't send this back.
								return
							}
							if err = item.Value(func(eventValue by) (err er) {
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
		defer quit.Q()
		for _, q := range queries {
			select {
			case <-r.Ctx.Done():
				return
			case <-c.Done():
				return
			default:
			}
			// search for the keys generated from the filter
			err = r.View(func(txn *badger.Txn) (err er) {
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
		err = r.View(func(txn *badger.Txn) (err er) {
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
				if chk.E(e) {
					err = nil
					continue
				}
				// send the event to client - the database stores correct JSON versions so no need to decode/encode.
				if _, err = fmt.Fprintf(w, "%s\n", b); chk.E(err) {
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
