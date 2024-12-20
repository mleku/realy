package app

import (
	"time"
	"realy.lol/kinds"
	"realy.lol/kind"
	"realy.lol/filter"
)

func (r *Relay) Spider() {
	// Don't start the spider if the spider key is not configured, many relays
	// require auth, whether the spider key is allowed, those that do,
	// and are, it must be there and this wraps the toggle together with the
	// configuration neatly.
	if len(r.C.SpiderKey) == 0 {
		return
	}
	// we run at first startup
	r.spider()
	// re-run the spider every hour to catch any updates that for whatever
	// reason permitted users may have uploaded to other relays via other
	// clients that may not be sending to us.
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-r.Ctx.Done():
			return
		case <-ticker.C:
			r.spider()
		}
	}
}

// RelayKinds are the types of events that we want to search and fetch.
var RelayKinds = &kinds.T{
	K: []*kind.T{
		kind.RelayListMetadata,
		kind.DMRelaysList,
	},
}

// spider is the actual function that does a spider run
func (r *Relay) spider() {
	// first find all the relays that we currently know about.
	filter := filter.T{Kinds: RelayKinds}
	sto := r.Storage()
	_ = filter
	_ = sto
}
