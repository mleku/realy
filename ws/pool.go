package ws

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/puzpuzpuz/xsync/v3"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/normalize"
	"realy.lol/signer"
	"realy.lol/timestamp"
)

var (
	seenAlreadyDropTick = 60
)

type SimplePool struct {
	Relays          *xsync.MapOf[string, *Client]
	Context         context.T
	authHandler     func() signer.I
	cancel          context.F
	eventMiddleware []func(IncomingEvent)
	// custom things not often used
	SignatureChecker func(*event.T) bool
}

type DirectedFilters struct {
	Filters *filters.T
	Client  string
}

type IncomingEvent struct {
	Event  *event.T
	Client *Client
}

func (ie IncomingEvent) String() string {
	return fmt.Sprintf("[%s] >> %s", ie.Client.URL, ie.Event.Serialize())
}

type PoolOption interface {
	ApplyPoolOption(*SimplePool)
}

func NewSimplePool(c context.T, opts ...PoolOption) *SimplePool {
	ctx, cancel := context.Cancel(c)

	pool := &SimplePool{
		Relays: xsync.NewMapOf[string, *Client](),

		Context: ctx,
		cancel:  cancel,
	}

	for _, opt := range opts {
		opt.ApplyPoolOption(pool)
	}

	return pool
}

// WithAuthHandler must be a function that signs the auth event when called.
// it will be called whenever any relay in the pool returns a `CLOSED` message
// with the "auth-required:" prefix, only once for each relay
type WithAuthHandler func() signer.I

func (h WithAuthHandler) ApplyPoolOption(pool *SimplePool) {
	pool.authHandler = h
}

// WithEventMiddleware is a function that will be called with all events received.
// more than one can be passed at a time.
type WithEventMiddleware func(IncomingEvent)

func (h WithEventMiddleware) ApplyPoolOption(pool *SimplePool) {
	pool.eventMiddleware = append(pool.eventMiddleware, h)
}

var (
	_ PoolOption = (WithAuthHandler)(nil)
	_ PoolOption = (WithEventMiddleware)(nil)
)

const MAX_LOCKS = 50

var namedMutexPool = make([]sync.Mutex, MAX_LOCKS)

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

func namedLock(name string) (unlock func()) {
	sptr := unsafe.StringData(name)
	idx := uint64(memhash(unsafe.Pointer(sptr), 0, uintptr(len(name)))) % MAX_LOCKS
	namedMutexPool[idx].Lock()
	return namedMutexPool[idx].Unlock
}

func (pool *SimplePool) EnsureRelay(url string) (*Client, error) {
	nm := string(normalize.URL(url))
	defer namedLock(nm)()

	relay, ok := pool.Relays.Load(nm)
	if ok && relay.IsConnected() {
		// already connected, unlock and return
		return relay, nil
	} else {
		var err error
		// we use this ctx here so when the pool dies everything dies
		ctx, cancel := context.Timeout(pool.Context, time.Second*15)
		defer cancel()

		opts := make([]RelayOption, 0, 1+len(pool.eventMiddleware))
		if pool.SignatureChecker != nil {
			opts = append(opts, WithSignatureChecker(pool.SignatureChecker))
		}

		if relay, err = RelayConnect(ctx, nm, opts...); chk.T(err) {
			return nil, errorf.E("failed to connect: %w", err)
		}

		pool.Relays.Store(nm, relay)
		return relay, nil
	}
}

// SubMany opens a subscription with the given filters to multiple relays
// the subscriptions only end when the context is canceled
func (pool *SimplePool) SubMany(c context.T, urls []string, ff *filters.T) chan IncomingEvent {
	return pool.subMany(c, urls, ff, true)
}

// SubManyNonUnique is like SubMany, but returns duplicate events if they come from different relays
func (pool *SimplePool) SubManyNonUnique(c context.T, urls []string, ff *filters.T) chan IncomingEvent {
	return pool.subMany(c, urls, ff, false)
}

func (pool *SimplePool) subMany(c context.T, urls []string, ff *filters.T,
	unique bool) chan IncomingEvent {
	ctx, cancel := context.Cancel(c)
	_ = cancel // do this so `go vet` will stop complaining
	events := make(chan IncomingEvent)
	seenAlready := xsync.NewMapOf[string, *timestamp.T]()
	ticker := time.NewTicker(time.Duration(seenAlreadyDropTick) * time.Second)
	eose := false
	pending := xsync.NewCounter()
	pending.Add(int64(len(urls)))
	for u, url := range urls {
		url = string(normalize.URL(url))
		urls[u] = url
		if idx := slices.Index(urls, url); idx != u {
			// skip duplicate relays in the list
			continue
		}

		go func(nm string) {
			var err error
			defer func() {
				pending.Dec()
				if pending.Value() == 0 {
					close(events)
				}
				cancel()
			}()
			hasAuthed := false
			interval := 3 * time.Second
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				var sub *Subscription
				var relay *Client
				if relay, err = pool.EnsureRelay(nm); chk.T(err) {
					goto reconnect
				}
				hasAuthed = false
			subscribe:
				if sub, err = relay.Subscribe(ctx, ff); chk.T(err) {
					goto reconnect
				}
				go func() {
					<-sub.EndOfStoredEvents
					eose = true
				}()
				// reset interval when we get a good subscription
				interval = 3 * time.Second

				for {
					select {
					case evt, more := <-sub.Events:
						if !more {
							// this means the connection was closed for weird reasons, like the server shut down
							// so we will update the filters here to include only events seem from now on
							// and try to reconnect until we succeed
							now := timestamp.Now()
							for i := range ff.F {
								ff.F[i].Since = now
							}
							goto reconnect
						}
						ie := IncomingEvent{Event: evt, Client: relay}
						for _, mh := range pool.eventMiddleware {
							mh(ie)
						}
						if unique {
							if _, seen := seenAlready.LoadOrStore(evt.EventId().String(),
								evt.CreatedAt); seen {
								continue
							}
						}
						select {
						case events <- ie:
						case <-ctx.Done():
						}
					case <-ticker.C:
						if eose {
							old := &timestamp.T{int64(timestamp.Now().Int() - seenAlreadyDropTick)}
							seenAlready.Range(func(id string, value *timestamp.T) bool {
								if value.I64() < old.I64() {
									seenAlready.Delete(id)
								}
								return true
							})
						}
					case reason := <-sub.ClosedReason:
						if strings.HasPrefix(reason,
							"auth-required:") && pool.authHandler != nil && !hasAuthed {
							// relay is requesting auth. if we can, we will perform auth and try again
							if err = relay.Auth(ctx, pool.authHandler()); err == nil {
								hasAuthed = true // so we don't keep doing AUTH again and again
								goto subscribe
							}
						} else {
							log.I.F("CLOSED from %s: '%s'\n", nm, reason)
						}
						return
					case <-ctx.Done():
						return
					}
				}
			reconnect:
				// we will go back to the beginning of the loop and try to connect again and again
				// until the context is canceled
				time.Sleep(interval)
				interval = interval * 17 / 10 // the next time we try we will wait longer
			}
		}(url)
	}

	return events
}

// SubManyEose is like SubMany, but it stops subscriptions and closes the channel when gets a EOSE
func (pool *SimplePool) SubManyEose(c context.T, urls []string, ff *filters.T) chan IncomingEvent {
	return pool.subManyEose(c, urls, ff, true)
}

// SubManyEoseNonUnique is like SubManyEose, but returns duplicate events if they come from different relays
func (pool *SimplePool) SubManyEoseNonUnique(c context.T, urls []string,
	ff *filters.T) chan IncomingEvent {
	return pool.subManyEose(c, urls, ff, false)
}

func (pool *SimplePool) subManyEose(c context.T, urls []string, ff *filters.T,
	unique bool) chan IncomingEvent {
	ctx, cancel := context.Cancel(c)

	events := make(chan IncomingEvent)
	seenAlready := xsync.NewMapOf[string, bool]()
	wg := sync.WaitGroup{}
	wg.Add(len(urls))

	go func() {
		// this will happen when all subscriptions get an eose (or when they die)
		wg.Wait()
		cancel()
		close(events)
	}()

	for _, url := range urls {
		go func(nm []byte) {
			var err error
			defer wg.Done()
			var client *Client
			if client, err = pool.EnsureRelay(string(nm)); chk.E(err) {
				return
			}

			hasAuthed := false

		subscribe:
			var sub *Subscription
			if sub, err = client.Subscribe(ctx, ff); chk.E(err) || sub == nil {
				log.E.F("error subscribing to %s with %v: %s", client, ff, err)
				return
			}
			for {
				select {
				case <-ctx.Done():
					return
				case <-sub.EndOfStoredEvents:
					return
				case reason := <-sub.ClosedReason:
					if strings.HasPrefix(reason,
						"auth-required:") && pool.authHandler != nil && !hasAuthed {
						// client is requesting auth. if we can we will perform auth and try again
						err := client.Auth(ctx, pool.authHandler())
						if err == nil {
							hasAuthed = true // so we don't keep doing AUTH again and again
							goto subscribe
						}
					}
					log.I.F("CLOSED from %s: '%s'\n", nm, reason)
					return
				case evt, more := <-sub.Events:
					if !more {
						return
					}

					ie := IncomingEvent{Event: evt, Client: client}
					for _, mh := range pool.eventMiddleware {
						mh(ie)
					}

					if unique {
						if _, seen := seenAlready.LoadOrStore(evt.EventId().String(),
							true); seen {
							continue
						}
					}

					select {
					case events <- ie:
					case <-ctx.Done():
						return
					}
				}
			}
		}(normalize.URL(url))
	}

	return events
}

// QuerySingle returns the first event returned by the first relay, cancels everything else.
func (pool *SimplePool) QuerySingle(c context.T, urls []string, f *filter.T) *IncomingEvent {
	ctx, cancel := context.Cancel(c)
	defer cancel()
	for ievt := range pool.SubManyEose(ctx, urls, filters.New(f)) {
		return &ievt
	}
	return nil
}

func (pool *SimplePool) batchedSubMany(
	c context.T,
	dfs []DirectedFilters,
	subFn func(context.T, []string, *filters.T, bool) chan IncomingEvent,
) chan IncomingEvent {
	res := make(chan IncomingEvent)

	for _, df := range dfs {
		go func(df DirectedFilters) {
			for ie := range subFn(c, []string{df.Client}, df.Filters, true) {
				res <- ie
			}
		}(df)
	}

	return res
}

// BatchedSubMany fires subscriptions only to specific relays, but batches them when they are the same.
func (pool *SimplePool) BatchedSubMany(c context.T, dfs []DirectedFilters) chan IncomingEvent {
	return pool.batchedSubMany(c, dfs, pool.subMany)
}

// BatchedSubManyEose is like BatchedSubMany, but ends upon receiving EOSE from all relays.
func (pool *SimplePool) BatchedSubManyEose(c context.T, dfs []DirectedFilters) chan IncomingEvent {
	return pool.batchedSubMany(c, dfs, pool.subManyEose)
}
