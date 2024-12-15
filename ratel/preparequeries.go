package ratel

import (
	"encoding/binary"
	"fmt"
	"math"

	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/timestamp"
	"realy.lol/ratel/prefixes"
)

type Results struct {
	Ev  *event.T
	TS  *timestamp.T
	Ser *serial.T
}

type query struct {
	index        no
	queryFilter  *filter.T
	searchPrefix by
	start        by
	skipTS       bo
}

// PrepareQueries analyses a filter and generates a set of query specs that produce
// key prefixes to search for in the badger key indexes.
func PrepareQueries(f *filter.T) (
	qs []query,
	ext *filter.T,
	since uint64,
	err er,
) {
	if f == nil {
		err = errorf.E("filter cannot be nil")
	}
	switch {
	// first if there is IDs, just search for them, this overrides all other filters
	case f.IDs.Len() > 0:
		qs = make([]query, f.IDs.Len())
		for i, idHex := range f.IDs.F() {
			ih := id.New(eventid.NewWith(by(idHex)))
			if ih == nil {
				log.E.F("failed to decode event ID: %s", idHex)
				// just ignore it, clients will be clients
				continue
			}
			prf := prefixes.Id.Key(ih)
			// log.T.F("id prefix to search on %0x from key %0x", prf, ih.Val)
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
			for i, pubkeyHex := range f.Authors.F() {
				var pk *pubkey.T
				if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
					// bogus filter, continue anyway
					continue
				}
				sp := prefixes.Pubkey.Key(pk)
				// log.I.F("search only for authors %0x from pub key %0x", sp, pk.Val)
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
			for _, pubkeyHex := range f.Authors.F() {
				for _, kind := range f.Kinds.K {
					var pk *pubkey.T
					if pk, err = pubkey.New(pubkeyHex); chk.E(err) {
						// skip this dodgy thing
						continue authors
					}
					ki := kinder.New(kind.K)
					sp := prefixes.PubkeyKind.Key(pk, ki)
					// log.T.F("search for authors from pub key %0x and kind %0x", pk.Val, ki.Val)
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
		for _, values := range f.Tags.Value() {
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
		// log.T.S(f.Tags.Value())
		for _, values := range f.Tags.Value() {
			// log.T.S(values.F())
			for _, value := range values.F()[1:] {
				// get key prefix (with full length) and offset where to write the last parts
				var prf by
				if prf, err = GetTagKeyPrefix(st(value)); chk.E(err) {
					continue
				}
				// remove the last part to get just the prefix we want here
				// log.T.F("search tags %0x", prf)
				qs[i] = query{index: i, queryFilter: f, searchPrefix: prf}
				i++
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
	default:
		if len(qs) > 0 {
			qs[0] = query{index: 0, queryFilter: f,
				searchPrefix: prefixes.CreatedAt.Key()}
			ext = nil
		}
		// log.T.S("other", qs)
	}
	var until int64 = math.MaxInt64
	if f.Until != nil {
		if fu := f.Until.I64(); fu < until {
			until = fu - 1
		}
	}
	for i, q := range qs {
		qs[i].start = binary.BigEndian.AppendUint64(q.searchPrefix, uint64(until))
	}
	// this is where we'll end the iteration
	if f.Since != nil {
		if fs := f.Since.U64(); fs > since {
			since = fs
		}
	}
	// if we got an empty filter, we still need a query for scraping the newest
	if len(qs) == 0 {
		qs = append(qs, query{index: 0, queryFilter: f, searchPrefix: by{1},
			start: by{1, 255, 255, 255, 255, 255, 255, 255, 255}})
	}
	return
}
