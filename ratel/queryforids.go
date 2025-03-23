package ratel

import (
	"realy.lol/context"
	"realy.lol/filter"
)

func (r *T) QueryForIds(c context.T, f *filter.T) (evids [][]byte, err error) {
	log.T.F("QueryForIds %s\n", f.Serialize())
	// evMap := make(map[string]*event.T)
	// var queries []query
	// var extraFilter *filter.T
	// var since uint64
	// if queries, extraFilter, since, err = PrepareQueries(f); chk.E(err) {
	// 	return
	// }

	return
}
