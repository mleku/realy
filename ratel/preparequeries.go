package ratel

import (
	"encoding/binary"
	"fmt"
	"math"

	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/log"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/timestamp"
)

type Results struct {
	Ev  *event.T
	TS  *timestamp.T
	Ser *serial.T
}

type query struct {
	index        int
	queryFilter  *filter.T
	searchPrefix []byte
	start        []byte
	skipTS       bool
}

// PrepareQueries analyses a filter and generates a set of query specs that produce
// key prefixes to search for in the badger key indexes.
func PrepareQueries(f *filter.T) (
	qs []query,
	ext *filter.T,
	since uint64,
	err error,
) {
	if f == nil {
		err = errorf.E("filter cannot be nil")
		return
	}
	switch {
	// first if there is IDs, just search for them, this overrides all other filters
	case f.IDs.Len() > 0:
		qs = make([]query, f.IDs.Len())
		for i, idB := range f.IDs.ToSliceOfBytes() {
			ih := id.New(eventid.NewWith(idB))
			if ih == nil {
				log.E.F("failed to decode event Id: %s", idB)
				// just ignore it, clients will be clients
				continue
			}
			prf := prefixes.Id.Key(ih)
			// log.T.ToSliceOfBytes("id prefix to search on %0x from key %0x", prf, ih.Val)
			qs[i] = query{
				index:        i,
				queryFilter:  f,
				searchPrefix: prf,
				skipTS:       true,
			}
		}
		// log.T.S("ids", qs)
		// second we make a set of queries based on author pubkeys, optionally with kinds
	case f.Authors.Len() > 0:
		// if there is no kinds, we just make the queries based on the author pub keys
		if f.Kinds.Len() == 0 {
			qs = make([]query, f.Authors.Len())
			for i, pubkeyHex := range f.Authors.ToSliceOfBytes() {
				var pk *pubkey.T
				if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
					// bogus filter, continue anyway
					continue
				}
				sp := prefixes.Pubkey.Key(pk)
				// log.I.ToSliceOfBytes("search only for authors %0x from pub key %0x", sp, pk.Val)
				qs[i] = query{
					index:        i,
					queryFilter:  f,
					searchPrefix: sp,
				}
			}
			// log.I.S("authors", qs)
		} else {
			// if there is kinds as well, we are searching via the kind/pubkey prefixes
			qs = make([]query, f.Authors.Len()*f.Kinds.Len())
			i := 0
		authors:
			for _, pubkeyHex := range f.Authors.ToSliceOfBytes() {
				for _, kind := range f.Kinds.K {
					var pk *pubkey.T
					if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
						// skip this dodgy thing
						continue authors
					}
					ki := kinder.New(kind.K)
					sp := prefixes.PubkeyKind.Key(pk, ki)
					// log.T.ToSliceOfBytes("search for authors from pub key %0x and kind %0x", pk.Val, ki.Val)
					qs[i] = query{index: i, queryFilter: f, searchPrefix: sp}
					i++
				}
			}
			// log.T.S("authors/kinds", qs)
		}
		if f.Tags.Len() > 0 {
			ext = &filter.T{Tags: f.Tags}
			// log.T.S("extra filter", ext)
		}
	case f.Tags.Len() > 0:
		// determine the size of the queries array by inspecting all tags sizes
		size := 0
		for _, values := range f.Tags.ToSliceOfTags() {
			size += values.Len() - 1
		}
		if size == 0 {
			return nil, nil, 0, fmt.Errorf("empty tag filters")
		}
		// we need a query for each tag search
		qs = make([]query, size)
		// and any kinds mentioned as well in extra filter
		ext = &filter.T{Kinds: f.Kinds}
		i := 0
		for _, values := range f.Tags.ToSliceOfTags() {
			tsb := values.ToSliceOfBytes()
			// indexable tags can only have 1 character in the key field.
			if len(tsb[0]) == 1 {
				for _, value := range tsb[1:] {
					// get key prefix (with full length) and offset where to write the last parts
					var prf []byte
					if prf, err = GetTagKeyPrefix(tsb[0][0], value); chk.E(err) {
						continue
					}
					// remove the last part to get just the prefix we want here
					qs[i] = query{index: i, queryFilter: f, searchPrefix: prf}
					i++
				}
			}
		}
		// log.T.S("tags", qs)
	case f.Kinds.Len() > 0:
		// if there is no ids, pubs or tags, we are just searching for kinds
		qs = make([]query, f.Kinds.Len())
		for i, kind := range f.Kinds.K {
			kk := kinder.New(kind.K)
			ki := prefixes.Kind.Key(kk)
			qs[i] = query{
				index:        i,
				queryFilter:  f,
				searchPrefix: ki,
			}
		}
	// log.T.S("kinds", qs)
	default: // todo: this is appearing on queries with only since/until
		log.I.F("nothing in filter, returning latest events")
		// if len(qs) > 0 {
		qs = append(qs, query{index: 0, queryFilter: f, searchPrefix: []byte{1},
			start: []byte{1, 255, 255, 255, 255, 255, 255, 255, 255},
			// })
			// qs = append(qs, query{index: 0, queryFilter: f,
			// 	searchPrefix: prefixes.CreatedAt.Key(),
			skipTS: true})
		ext = nil
		// }
		// 	// log.T.S("other", qs)
	}

	// this is where we'll end the iteration
	if f.Since != nil {
		if fs := f.Since.U64(); fs > since {
			since = fs
		}
	}
	// log.I.ToSliceOfBytes("since %d", since)

	var until uint64 = math.MaxInt64
	if f.Until != nil {
		if fu := f.Until.U64(); fu < until {
			until = fu + 1
		}
	}
	// log.I.ToSliceOfBytes("until %d", until)
	for i, q := range qs {
		qs[i].start = binary.BigEndian.AppendUint64(q.searchPrefix, uint64(until))
	}
	// if we got an empty filter, we still need a query for scraping the newest
	if len(qs) == 0 {
		qs = append(qs, query{index: 0, queryFilter: f, searchPrefix: []byte{1},
			start: []byte{1, 255, 255, 255, 255, 255, 255, 255, 255}})
	}
	return
}
