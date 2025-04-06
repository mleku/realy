// Package layer2 is a library for building nostr event stores with two separate
// data storage systems, primarily for creating size limited caches with larger
// stores backing them, to enable scaling providing access to an event store to
// more users more quickly via a caching strategy.
package layer2

import (
	"errors"
	"io"
	"path/filepath"
	"sync"
	"time"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/timestamp"
)

// Backend is a two level nostr event store. The first level is assumed to have a subset of all
// events that the second level has. This is a mechanism for sharding nostr event data across
// multiple relays which can then be failovers for each other or shards by geography or subject
// matter.
type Backend struct {
	Ctx  context.T
	WG   *sync.WaitGroup
	path string
	// L1 will store its state/configuration in path/layer1
	L1 store.I
	// L2 will store its state/configuration in path/layer2
	L2 store.I
	// PollFrequency is how often the L2 is queried for recent events. This is only
	// relevant for shared layer2 stores, and will not apply for layer2
	// implementations that are just two separate data store systems on the same
	// server.
	PollFrequency time.Duration
	// PollOverlap is the multiple of the PollFrequency within which polling the L2
	// is done to ensure any slow synchrony on the L2 is covered (2-4 usually).
	PollOverlap int
	// EventSignal triggers when the L1 saves a new event from the L2
	//
	// caller is responsible for populating this so that a signal can pass to all
	// peers sharing the same L2 and enable cross-cluster subscription delivery.
	EventSignal event.C
}

// Init a layer2.Backend setting up their configurations and polling frequencies and other
// similar things.
func (b *Backend) Init(path string) (err error) {
	b.path = path
	// each backend will have configuration files living in a subfolder of the same
	// root, path/layer1 and path/layer2 - this may only be state/configuration, or
	// it can be the site of the storage of data.
	path1 := filepath.Join(path, "layer1")
	path2 := filepath.Join(path, "layer2")
	if err = b.L1.Init(path1); chk.E(err) {
		return
	}
	if err = b.L2.Init(path2); chk.E(err) {
		return
	}
	// if poll syncing is disabled don't start the ticker
	if b.PollFrequency == 0 {
		return
	}
	// Polling overlap should be 4x polling frequency, if less than 2x
	if b.PollOverlap < 2 {
		b.PollOverlap = 4
	}
	log.I.Ln("L2 polling frequency", b.PollFrequency, "overlap",
		b.PollFrequency*time.Duration(b.PollOverlap))
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		last := timestamp.Now().I64()
	out:
		for {
			select {
			case <-b.Ctx.Done():
				chk.E(b.Close())
				return
			case <-ticker.C:
				until := timestamp.Now()
				var evs []*event.T
				if evs, err = b.L2.QueryEvents(b.Ctx,
					&filter.T{Since: timestamp.FromUnix(last), Until: until}); chk.E(err) {
					continue out
				}
				// todo now wat
				_ = evs
				last = until.I64() - int64(time.Duration(b.PollOverlap)*b.PollFrequency/time.Second)
			}
		}
	}()
	return
}

// Path returns the filesystem path root of the layer2.Backend.
func (b *Backend) Path() (s string) { return b.path }

// Close the two layers of a layer2.Backend.
func (b *Backend) Close() (err error) {
	var e1, e2 error
	if e1 = b.L1.Close(); chk.E(e1) {
		err = e1
	}
	if e2 = b.L2.Close(); chk.E(e2) {
		if err != nil {
			err = errors.Join(err, e2)
		} else {
			err = e2
		}
	}
	return
}

// Nuke wipes the both of the event stores in parallel and returns when both are complete.
func (b *Backend) Nuke() (err error) {
	var wg sync.WaitGroup
	var err1, err2 error
	go func() {
		if err1 = b.L1.Nuke(); chk.E(err) {
		}
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		if err2 = b.L2.Nuke(); chk.E(err) {
		}
		wg.Done()
	}()
	wg.Wait()
	err = errors.Join(err1, err2)
	return
}

// QueryEvents processes a filter.T search on the event store. The events found in the second
// level will be saved into the first level so they become available from the first layer next
// time they match.
func (b *Backend) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err error) {
	if evs, err = b.L1.QueryEvents(c, f); chk.E(err) {
		return
	}
	// if there is pruned events (have only Id, no pubkey), they will also be in the
	// L2 result, save these to the L1.
	var revives [][]byte
	var founds event.Ts
	for _, ev := range evs {
		if len(ev.Pubkey) == 0 {
			// note the event Id to fetch
			revives = append(revives, ev.Id)
		} else {
			founds = append(founds, ev)
		}
	}
	evs = founds
	go func(revives [][]byte) {
		var err error
		// construct the filter to fetch the missing events in the background that we
		// know about, these will come in later on the subscription while it remains
		// open.
		l2filter := &filter.T{IDs: tag.New(revives...)}
		var evs2 event.Ts
		if evs2, err = b.L2.QueryEvents(c, l2filter); chk.E(err) {
			return
		}
		for _, ev := range evs2 {
			// saving the events here will trigger a match on the subscription
			if err = b.L1.SaveEvent(c, ev); err != nil {
				continue
			}
		}
		// after fetching what we know exists of non pruned indexes that found stubs we
		// want to run the query to the L2 anyway, and any matches that are found that
		// were not locally available will now be available.
		//
		// if the subscription is still open the matches will be delivered later, the
		// late events will be in descending (reverse chronological) order but the stream
		// as a whole will not be. whatever.
		var evs event.Ts
		if evs, err = b.L2.QueryEvents(c, f); chk.E(err) {
			return
		}
		for _, ev := range evs {
			if err = b.L1.SaveEvent(c, ev); err != nil {
				continue
			}
		}
	}(revives)
	return
}

// // CountEvents counts how many events match on a filter, providing an approximate flag if either
// // of the layers return this, and the result is the maximum of the two layers results.
// func (b *Backend) CountEvents(c context.T, f *filter.T) (count int, approx bool, err error) {
// 	var wg sync.WaitGroup
// 	var count1, count2 int
// 	var approx1, approx2 bool
// 	var err1, err2 error
// 	go func() {
// 		count1, approx1, err1 = b.L1.CountEvents(c, f)
// 		wg.Done()
// 	}()
// 	// because this is a low-data query we will wait until the L2 also gets a count,
// 	// which should be under a few hundred ms in most cases
// 	go func() {
// 		wg.Add(1)
// 		count2, approx2, err2 = b.L2.CountEvents(c, f)
// 	}()
// 	wg.Wait()
// 	// we return the maximum, it is assumed the L2 is authoritative, but it could be
// 	// the L1 has more for whatever reason, so return the maximum of the two.
// 	count = count1
// 	approx = approx1
// 	if count2 > count {
// 		count = count2
// 		// the approximate flag probably will be false if the L2 got more, and it is a
// 		// very large, non GC store.
// 		approx = approx2
// 	}
// 	err = errors.Join(err1, err2)
// 	return
// }

// DeleteEvent deletes an event on both the layer1 and layer2.
func (b *Backend) DeleteEvent(c context.T, ev *eventid.T, noTombstone ...bool) (err error) {
	// delete the events from both stores.
	err = errors.Join(b.L1.DeleteEvent(c, ev, noTombstone...),
		b.L2.DeleteEvent(c, ev, noTombstone...))
	return
}

// SaveEvent stores an event on both layer1 and layer2.
func (b *Backend) SaveEvent(c context.T, ev *event.T) (err error) {
	// save to both event stores
	err = errors.Join(
		b.L1.SaveEvent(c, ev), // this will also send out to subscriptions
		b.L2.SaveEvent(c, ev))
	return
}

// Import events to the layer2, if the events come up in searches they will be propagated down
// to the layer1.
func (b *Backend) Import(r io.Reader) {
	// we import up to the L2 directly, demanded data will be fetched from it by
	// later queries.
	b.L2.Import(r)
}

// Export from the layer2, which is assumed to be the most authoritative (and large) store of
// events available to the relay.
func (b *Backend) Export(c context.T, w io.Writer, pubkeys ...[]byte) {
	// export only from the L2 as it is considered to be the authoritative event
	// store of the two, and this is generally an administrative or infrequent action
	// and latency will not matter as it usually will be a big bulky download.
	b.L2.Export(c, w, pubkeys...)
}

// Sync triggers both layer1 and layer2 to flush their buffers and store any events in caches.
func (b *Backend) Sync() (err error) {
	err1 := b.L1.Sync()
	// more than likely L2 sync is a noop.
	err2 := b.L2.Sync()
	err = errors.Join(err1, err2)
	return
}
