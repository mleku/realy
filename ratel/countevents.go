package ratel

import (
	"realy.lol/event"
	"realy.lol/filter"
)

func (r *T) CountEvents(c Ctx, f *filter.T) (count N, err E) {
	var evs []*event.T
	evs, err = r.QueryEvents(c, f)
	count = len(evs)
	evs = nil
	return
}
