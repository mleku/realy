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
	"realy.lol/ratel/keys/lang"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/store"
	"realy.lol/tag"
)

type FulltextSequence struct {
	distinctMatches    int
	distinctCount      int
	distance, sequence float64
	items              []*prefixes.FulltextIndexKey
}

func (r *T) QueryFulltextEvents(c context.T, f *filter.T) (evs []store.IdTsPk, err error) {
	start := time.Now()
	// just use QueryEvents if there isn't actually any fulltext search field content.
	if len(f.Search) == 0 {
		return r.QueryForIds(c, f)
	}
	split := bytes.Split(f.Search, []byte{' '})
	var lTag []byte
	var terms [][]byte
	for i := range split {
		if bytes.HasPrefix(split[i], []byte("lang:")) {
			lTag = split[i][5:]
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
				if len(lTag) > 0 {
					if _, err = txn.Get(prefixes.LangIndex.Key(lang.New(lTag), ser)); chk.E(err) {
						// the event does not have an associated language tag
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
	// get the number of distinct terms that were matched, and the counts of each term
	var distinct map[string]int
	for _, g := range groupS {
		for _, i := range g.items {
			if _, ok := distinct[string(i.Word())]; ok {
				distinct[string(i.Word())]++
			} else {
				distinct[string(i.Word())] = 1
			}
		}
		g.distinctMatches = len(distinct)
		// sum the counts on the distinct matches to get the number of distinct match groups
		for _, dCount := range distinct {
			g.distinctCount += dCount
		}
	}
	// eliminate all results that don't have a complete set
	var tmp []FulltextSequence
	for _, g := range groupS {
		if g.distinctCount >= len(g.items) {
			tmp = append(tmp, g)
		}
	}
	groupS = tmp
	// sort each group so the indexes are sorted by sequence number
	for _, g := range groupS {
		sort.Slice(g.items, func(i, j int) bool {
			return g.items[i].Sequence().Val < g.items[j].Sequence().Val
		})
	}
	// find the max sequential in the items
	for _, g := range groupS {
		var itemSequence int
		var itemsInSequence, distanceInSequence [][]int
		var seq, dist []int
		var groupSeq, groupDist []float64
		for itemSequence < len(g.items) {
			for _, ts := range terms {
				if bytes.Equal(g.items[itemSequence].Word(), ts) {
					seq = append(seq, itemSequence)
					dist = append(dist,
						int(g.items[itemSequence].Sequence().Val))
					itemSequence++
				} else {
					itemsInSequence = append(itemsInSequence, seq)
					distanceInSequence = append(distanceInSequence, dist)
					dist = []int{}
					seq = []int{}
					itemSequence++
					continue
				}
			}
		}
		// generate the sequence and distance scores for the groups
		for i := range itemsInSequence {
			// add the number of items in sequence to an array
			groupSeq = append(groupSeq, float64(len(itemsInSequence[i])))
			// add the distance between the first and last items in a sequence to an array
			groupDist = append(groupDist,
				float64(distanceInSequence[i][len(distanceInSequence)-1]-
					distanceInSequence[i][0]))
		}
		// sum the sequence and distance numbers for averaging
		var s, d float64
		for i := range groupSeq {
			s += groupSeq[i]
			d += groupDist[i]
		}
		// get the average sequence and distance values
		s /= float64(len(groupSeq))
		d /= float64(len(groupDist))
		// divide by the length of the terms
		s /= float64(len(terms))
		d /= float64(len(terms))
		// these values represent sequence ond proximity, store in the group
		g.sequence = s
		g.distance = d
	}
	// sort the groups by these scores by ascending distance and descending sequence
	sort.Slice(groupS, func(i, j int) bool {
		return groupS[i].distance < groupS[j].distance &&
			groupS[i].sequence > groupS[j].sequence
	})
	if f.Limit != nil {
		evs = evs[:*f.Limit]
	} else {
		evs = evs[:r.MaxLimit]
	}
	log.I.F("performed search for '%s' in %v", f.Search, time.Now().Sub(start))
	return
}
