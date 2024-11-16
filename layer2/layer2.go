package layer2

import (
	"errors"
	"io"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/store"
	"realy.lol/timestamp"
	"sync"
	"time"
)

type Backend struct {
	Ctx  context.T
	WG   *sync.WaitGroup
	path S
	L1   store.I
	L2   store.I
	// PollFrequency is how often the L2 is queried for recent events
	PollFrequency time.Duration
	// PollOverlap is the multiple of the PollFrequency within which polling the L2
	// is done to ensure any slow synchrony on the L2 is covered (2-4 usually)
	PollOverlap timestamp.T
	// EventSignal triggers when the L1 saves a new event from the L2
	//
	// caller is responsible for populating this
	EventSignal event.C
}

func (b *Backend) Init(path S) (err E) {
	b.path = path
	if err = b.L1.Init(path); chk.E(err) {
		return
	}
	if err = b.L2.Init(path); chk.E(err) {
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
				b.Close()
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
				last = until.I64() - b.PollOverlap.I64()*int64(b.PollFrequency/time.Second)
			}
		}
	}()
	return
}

func (b *Backend) Path() (s S) { return b.path }

func (b *Backend) Close() (err E) {
	var e1, e2 E
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

func (b *Backend) Nuke() (err E) {
	var wg sync.WaitGroup
	var err1, err2 E
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

func (b *Backend) QueryEvents(c Ctx, f *filter.T) (evs []*event.T, err E) {
	var evs1, evs2 []*event.T
	var err1, err2 E
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		if evs2, err1 = b.L2.QueryEvents(c, f); chk.E(err) {
		}
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		if evs1, err2 = b.L1.QueryEvents(c, f); chk.E(err) {
		}
		wg.Done()
	}()
	wg.Wait()
	// both or either will return if the context is closed
	evs = append(evs, evs1...)
	evs = append(evs, evs2...)
	err = errors.Join(err1, err2)
	return
}

func (b *Backend) CountEvents(c Ctx, f *filter.T) (count N, approx bool, err E) {
	var wg sync.WaitGroup
	var count1, count2 N
	var approx1, approx2 bool
	var err1, err2 E
	go func() {
		count1, approx1, err1 = b.L1.CountEvents(c, f)
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		count2, approx2, err2 = b.L2.CountEvents(c, f)
	}()
	wg.Wait()
	// we return the maximum, it is assumed the L2 is authoritative, but it could be
	// the L1 has more for whatever reason, so return the maximum of the two.
	count = count1
	if count2 > count {
		count = count2
	}
	err = errors.Join(err1, err2)
	// if either are approximate, we mark the result approximate.
	approx = approx1 || approx2
	return
}

func (b *Backend) DeleteEvent(c Ctx, ev *eventid.T) (err E) {
	// delete the events
	err = errors.Join(b.L1.DeleteEvent(c, ev), b.L2.DeleteEvent(c, ev))
	return
}

func (b *Backend) SaveEvent(c Ctx, ev *event.T) (err E) {
	err = errors.Join(b.L1.SaveEvent(c, ev), b.L2.SaveEvent(c, ev))
	return
}

func (b *Backend) Import(r io.Reader) {
	// for an L2 we want to put only to the L1 and push to L2 when it gets stale
	b.L1.Import(r)
}

func (b *Backend) Export(c Ctx, w io.Writer, pubkeys ...B) {
	// do this in series, local first. L2 may not even have an export function.
	// todo: sorta seems like maybe L2 is authoritative and don't need to export L1? deduplicating will be expensive.
	b.L1.Export(c, w, pubkeys...)
	b.L2.Export(c, w, pubkeys...)
}

func (b *Backend) Sync() (err E) {
	err1 := b.L1.Sync()
	// more than likely L2 sync is a noop.
	err2 := b.L2.Sync()
	err = errors.Join(err1, err2)
	return
}
