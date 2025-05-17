package ratel

import (
	"bytes"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/filter"
	"realy.lol/log"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/store"
	"realy.lol/tag"
)

type FulltextSequence struct {
	inSequence int
	distance   int
	sequence   []int
	items      []*prefixes.FulltextIndexKey
}

func (r *T) QueryFulltextEvents(c context.T, f *filter.T) (evs []store.IdTsPk, err error) {
	start := time.Now()
	// just use QueryEvents if there isn't actually any fulltext search field content.
	if len(f.Search) == 0 {
		return r.QueryForIds(c, f)
	}
	split := bytes.Split(f.Search, []byte{' '})
	var lang []byte
	var terms [][]byte
	for i := range split {
		if bytes.HasPrefix(split[i], []byte("lang:")) {
			lang = split[i][5:]
		} else {
			terms = append(terms, split[i])
		}
	}
	var fTags []*tag.T
	if f.Tags != nil {
		fTags = f.Tags.ToSliceOfTags()
	}
	fAut := f.Authors.ToSliceOfBytes()
	fKinds := f.Kinds.K
	var matches []*prefixes.FulltextIndexKey
	if err = r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{
			Prefix:  prefixes.FulltextIndex.Key(),
			Reverse: true,
		})
		defer it.Close()
		for _, v := range terms {
			for it.Rewind(); it.ValidForPrefix(prefixes.FulltextIndex.Key(arb.New(v))); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)
				var idx *prefixes.FulltextIndexKey
				if idx, err = prefixes.NewFulltextIndexKey(k); chk.E(err) {
					continue
				}
				if f.Since != nil {
					ts := idx.Timestamp()
					if ts.I64() < f.Since.I64() {
						// event is earlier than since
						continue
					}
				}
				if f.Until != nil {
					ts := idx.Timestamp()
					if ts.I64() > f.Until.I64() {
						// event is later than until
						continue
					}
				}
				if len(fKinds) != 0 {
					var found bool
					ki := idx.Kind()
					for _, kin := range fKinds {
						if ki.Equal(kin) {
							found = true
							break
						}
					}
					// kinds are present in filter and don't match
					if !found {
						continue
					}
				}
				if len(fAut) > 0 {
					var found bool
					pk := idx.Pubkey()
					for _, p := range fAut {
						if bytes.Equal(p, pk) {
							found = true
							break
						}
					}
					// pubkey is in filter and doesn't match
					if !found {
						continue
					}
				}
				// get serial
				ser := idx.Serial()
				// check language tags
				if len(lang) > 0 {
					var found bool
					func() {
						itl := txn.NewIterator(badger.IteratorOptions{
							Prefix: prefixes.LangIndex.Key(),
						})
						defer itl.Close()
						for itl.Rewind(); itl.Valid(); itl.Next() {
							s := serial.FromKey(itl.Item().KeyCopy(nil))
							if s.Uint64() == ser.Uint64() {
								found = true
								return
							}
						}
					}()
					// the event does not have an associated language tag
					if !found {
						continue
					}
				}
				// now we can check tags, they can't be squished into a fulltext index, and
				// require a second table iteration
				if len(fTags) > 0 {
					var found bool
					for _, ft := range fTags {
						if len(ft.Key()) == 2 && ft.Key()[0] == '#' {
							var tp []byte
							if tp, err = GetTagKeyPrefix(ft.Key()[0], ft.Value()); chk.E(err) {
								continue
							}
							if len(tp) == 0 {
								// the tag did not generate an index
								continue
							}
							func() {
								itt := txn.NewIterator(badger.IteratorOptions{
									Prefix: tp,
								})
								defer itt.Close()
								for itt.Rewind(); itt.Valid(); itt.Next() {
									s := serial.FromKey(itt.Item().KeyCopy(nil))
									if s.Uint64() == ser.Uint64() {
										found = true
										return
									}
								}
							}()
							// the event does not have any of the required tags
							if !found {
								continue
							}
						}
					}
					if !found {
						continue
					}
				}
				// if we got to here, we have a match
				matches = append(matches, idx)
			}
		}
		return
	}); chk.E(err) {
		return
	}
	if len(matches) == 0 {
		// didn't find any (?)
		return
	}
	// next we need to group and sort the results
	groups := make(map[uint64]FulltextSequence)
	for _, v := range matches {
		if _, ok := groups[v.Serial().Uint64()]; !ok {
			groups[v.Serial().Uint64()] = FulltextSequence{items: []*prefixes.FulltextIndexKey{v}}
		} else {
			g := groups[v.Serial().Uint64()]
			g.items = append(g.items, v)
		}
	}
	// now we need to convert the map to a slice so we can sort it
	var groupS []FulltextSequence
	for _, g := range groups {
		groupS = append(groupS, g)
	}
	// first, sort the groups by the number of elements in descending order
	sort.Slice(groupS, func(i, j int) (e bool) {
		return len(groupS[i].items) > len(groupS[j].items)
	})
	// get the distance of the groups
	for _, g := range groupS {
		g.distance = int(g.items[len(g.items)-1].Sequence().Val - g.items[0].Sequence().Val)
	}
	// get the sequence as relates to the search terms
	for _, g := range groupS {
		for i := range g.items {
			if i > 0 {
				for k := range terms {
					if bytes.Equal(g.items[i].Word(), terms[k]) {
						g.sequence = append(g.sequence, i)
					}
				}
			}
		}
	}
	// count the number of elements of the sequence that are in ascending order
	for _, g := range groupS {
		for i := range g.sequence {
			if i > 0 {
				if g.sequence[i-1] < g.sequence[i] {
					g.inSequence++
				}
			}
		}
	}
	// find the boundaries of each length segment of the group
	var groupedCounts []int
	var lastCount int
	lastCount = len(groupS[0].items)
	for i, g := range groupS {
		if len(g.items) < lastCount {
			groupedCounts = append(groupedCounts, i)
			lastCount = len(g.items)
		}
	}
	// break the groupS into segments of the same length
	var segments [][]FulltextSequence
	lastCount = 0
	for i := range groupedCounts {
		segments = append(segments, groupS[lastCount:groupedCounts[i]])
	}
	// sort the segments by distance and number in sequence
	for _, s := range segments {
		sort.Slice(s, func(i, j int) bool {
			return (s[i].distance < s[j].distance) && s[i].inSequence > s[i].inSequence
		})
	}
	// flatten the segments back into a list
	var list []FulltextSequence
	for _, seg := range segments {
		for _, bit := range seg {
			list = append(list, bit)
		}
	}
	// convert into store.IdTsPk
	for _, bit := range list {
		for _, el := range bit.items {
			evs = append(evs, store.IdTsPk{
				Ts:  el.Timestamp().I64(),
				Id:  el.EventId().Bytes(),
				Pub: el.Pubkey(),
			})
		}
	}
	if f.Limit != nil {
		evs = evs[:*f.Limit]
	} else {
		evs = evs[:r.MaxLimit]
	}
	log.I.F("performed search for '%s' in %v", f.Search, time.Now().Sub(start))
	return
}
