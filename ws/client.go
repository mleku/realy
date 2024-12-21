// Package ws is a websocket library
//
// todo: this client code is bullshit
package ws

import (
	"bytes"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/puzpuzpuz/xsync/v3"

	"realy.lol/atomic"
	"realy.lol/auth"
	"realy.lol/context"
	"realy.lol/envelopes"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closedenvelope"
	"realy.lol/envelopes/countenvelope"
	"realy.lol/envelopes/eoseenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/noticeenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/kind"
	"realy.lol/normalize"
	"realy.lol/signer"
	"realy.lol/qu"
)

type Status no

var subscriptionIDCounter atomic.Int64

type Client struct {
	closeMutex                    sync.Mutex
	URL                           st
	RequestHeader                 http.Header // e.g. for origin header
	Connection                    *Connection
	Subscriptions                 *xsync.MapOf[st, *Subscription]
	ConnectionError               er
	connectionContext             cx // will be canceled when the connection closes
	connectionContextCancel       context.F
	challenge                     by // NIP-42 challenge, we only keep the last
	noticeHandler                 func(st)
	customHandler                 func(by)
	okCallbacks                   *xsync.MapOf[st, func(bo, st)]
	writeQueue                    chan writeRequest
	subscriptionChannelCloseQueue chan *Subscription
}

type writeRequest struct {
	msg    by
	answer chan er
}

// NewRelay returns a new relay. The relay connection will be closed when the context is canceled.
func NewRelay(c cx, url st) *Client {
	ctx, cancel := context.Cancel(c)
	r := &Client{
		URL:                           st(normalize.URL(by(url))),
		connectionContext:             ctx,
		connectionContextCancel:       cancel,
		Subscriptions:                 xsync.NewMapOf[st, *Subscription](),
		okCallbacks:                   xsync.NewMapOf[st, func(bo, st)](),
		writeQueue:                    make(chan writeRequest),
		subscriptionChannelCloseQueue: make(chan *Subscription),
	}
	return r
}

// RelayConnect returns a relay object connected to url. Once successfully connected, cancelling
// ctx has no effect. To close the connection, call r.Close().
func RelayConnect(ctx cx, url st) (*Client, er) {
	r := NewRelay(context.Bg(), url)
	err := r.Connect(ctx)
	return r, err
}

// String just returns the relay URL.
func (r *Client) String() st { return r.URL }

// Context retrieves the context that is associated with this relay connection.
func (r *Client) Context() cx { return r.connectionContext }

// IsConnected returns true if the connection to this relay seems to be active.
func (r *Client) IsConnected() bo { return r.connectionContext.Err() == nil }

// Connect tries to establish a websocket connection to r.URL. If the context
// expires before the connection is complete, an error is returned. Once
// successfully connected, context expiration has no effect: call r.Close to
// close the connection.
//
// The underlying relay connection will use a background context. If you want to
// pass a custom context to the underlying relay connection, use NewRelay() and
// then Client.Connect().
func (r *Client) Connect(c cx) er { return r.ConnectWithTLS(c, nil) }

// ConnectWithTLS tries to establish a secured websocket connection to r.URL
// using customized tls.Config (CA's, etc.).
func (r *Client) ConnectWithTLS(ctx cx, tlsConfig *tls.Config) (err er) {
	if r.connectionContext == nil || r.Subscriptions == nil {
		return errorf.E("relay must be initialized with a call to NewRelay()")
	}
	if r.URL == "" {
		return errorf.E("relay url unset")
	}
	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		ctx, cancel = context.Timeout(ctx, 7*time.Second)
		defer cancel()
	}
	if r.RequestHeader != nil && r.RequestHeader.Get("User-Agent") == "" {
		r.RequestHeader.Set("User-Agent", "realy.lol")
	}
	if r.Connection, err = NewConnection(ctx, r.URL, r.RequestHeader,
		tlsConfig); chk.E(err) {

		return errorf.E("error opening websocket to '%s': %w",
			r.URL, err)
	}
	// ping every 29 seconds (??)
	ticker := time.NewTicker(29 * time.Second)
	// to be used when the connection is closed
	go func() {
		<-r.connectionContext.Done()
		// stop the ticker
		ticker.Stop()
		r.Connection = nil
		// close all subscriptions
		for _, sub := range r.Subscriptions.Range {
			sub.Unsub()
		}
	}()
	// queue all write operations here so we don't do mutex spaghetti
	go func() {
		var err er
		for {
			select {
			case <-ticker.C:
				if r.Connection != nil {
					if err = wsutil.WriteClientMessage(r.Connection.conn,
						ws.OpPing, nil); chk.E(err) {

						log.D.F("client ( %s ) error writing ping: %v; "+
							"closing websocket", r.URL, err)
						chk.E(r.Close()) // this should cancel the context
						return
					}
				}
			case wr := <-r.writeQueue:
				// all write requests will go through this to prevent races
				if err = r.Connection.
					WriteMessage(r.connectionContext, wr.msg); chk.T(err) {

					wr.answer <- err
				}
				close(wr.answer)
			case <-r.connectionContext.Done():
				// stop here
				return
			}
		}
	}()
	// general message reader loop
	go func() {
		for {
			buf := new(bytes.Buffer)
			// buf.Reset()
			if err := r.Connection.
				ReadMessage(r.connectionContext, buf); chk.T(err) {

				r.ConnectionError = err
				chk.E(r.Close())
				break
			}
			msg := buf.Bytes()
			log.T.F("client ( %s ) <- %s", r.URL, msg)

			var t st
			if t, msg, err = envelopes.Identify(msg); chk.E(err) {
				continue
			}

			var rem by
			switch t {
			case noticeenvelope.L:
				env := noticeenvelope.New()
				if env, msg, err = noticeenvelope.Parse(msg); chk.E(err) {
					continue
				}
				log.E.F("NOTICE from %s: '%s'\n", r.URL, env.Message)

			case authenvelope.L:
				env := authenvelope.NewChallenge()
				if env, msg, err = authenvelope.ParseChallenge(msg); chk.E(err) {
					continue
				}
				if len(env.Challenge) == 0 {
					continue
				}
				r.challenge = env.Challenge

			case eventenvelope.L:
				env := eventenvelope.NewResult()
				if rem, err = env.Unmarshal(msg); chk.E(err) {
					continue
				}
				if len(rem) > 0 {
					log.I.S(rem)
				}
				if len(env.Subscription.T) == 0 {
					continue
				}
				if sub, ok := r.Subscriptions.
					Load(env.Subscription.String()); !ok {

					log.D.F("{%s} no subscription with id '%s'\n",
						r.URL, env.Subscription)
					continue

				} else {
					// check if the event matches the desired filter, ignore
					// otherwise
					if !sub.Filters.Match(env.Event) {
						log.D.F("{%s} filter does not match: %v ~ %v\n",
							r.URL, sub.Filters, env.Event)
						continue
					}
					// dispatch this to the internal events channel of the
					// subscription
					sub.dispatchEvent(env.Event)
				}

			case eoseenvelope.L:
				var env *eoseenvelope.T
				if env, rem, err = eoseenvelope.Parse(msg); chk.E(err) {
					continue
				}
				if subscription, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
					subscription.dispatchEose()
				}

			case closedenvelope.L:
				var env *closedenvelope.T
				if env, rem, err = closedenvelope.Parse(msg); chk.E(err) {
					continue
				}
				if subscription, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
					subscription.dispatchClosed(env.ReasonString())
				}

			case countenvelope.L:
				var env *countenvelope.Response
				if env, rem, err = countenvelope.Parse(msg); chk.E(err) {
					continue
				}
				if subscription, ok := r.Subscriptions.Load(env.ID.String()); ok && subscription.countResult != nil {
					subscription.countResult <- env.Count
				}

			case okenvelope.L:
				var env *okenvelope.T
				if env, rem, err = okenvelope.Parse(msg); chk.E(err) {
					continue
				}
				if cb, ok := r.okCallbacks.Load(env.EventID.String()); ok {
					cb(env.OK, env.ReasonString())
				} else {
					log.I.F("{%s} got an unexpected OK message for event %s",
						r.URL, env.EventID)
				}
			}
		}
	}()
	return nil
}

// Write queues a message to be sent to the relay.
func (r *Client) Write(msg by) <-chan er {
	ch := make(chan er)
	select {
	case r.writeQueue <- writeRequest{msg: msg, answer: ch}:
	case <-r.connectionContext.Done():
		go func() { ch <- errorf.E("connection closed") }()
	}
	return ch
}

// Publish sends an "EVENT" command to the relay r as in NIP-01 and waits for an OK response.
func (r *Client) Publish(c cx, ev *event.T) er { return r.publish(c, ev) }

// Auth sends an "AUTH" command client->relay as in NIP-42 and waits for an OK response.
func (r *Client) Auth(c cx, sign signer.I) (err er) {
	authEvent := auth.CreateUnsigned(sign.Pub(), r.challenge, r.URL)
	if err = authEvent.Sign(sign); chk.T(err) {
		return errorf.E("error signing auth event: %w", err)
	}
	return r.publish(c, authEvent)
}

// publish can be used both for EVENT and for AUTH
func (r *Client) publish(ctx cx, ev *event.T) (err er) {
	var cancel context.F
	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		ctx, cancel = context.TimeoutCause(ctx, 7*time.Second,
			errorf.E("given up waiting for an OK"))
		defer cancel()
	} else {
		// otherwise make the context cancellable, so we can stop everything
		// upon receiving an "OK"
		ctx, cancel = context.Cancel(ctx)
		defer cancel()
	}
	// listen for an OK callback
	var gotOk bo
	id := ev.IDString()
	r.okCallbacks.Store(id, func(ok bo, reason st) {
		gotOk = true
		if !ok {
			err = errorf.E("msg: %s", reason)
		}
		cancel()
	})
	defer r.okCallbacks.Delete(id)
	// publish event
	var b by
	if ev.Kind.Equal(kind.ClientAuthentication) {
		if b = authenvelope.NewResponseWith(ev).Marshal(b); chk.E(err) {
			return
		}
	} else {
		if b = eventenvelope.NewSubmissionWith(ev).Marshal(b); chk.E(err) {
			return
		}
	}
	log.T.F("{%s} sending %s\n", r.URL, b)
	if err = <-r.Write(b); chk.T(err) {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			// this will be called when we get an OK or when the context has
			// been canceled
			if gotOk {
				return err
			}
			return ctx.Err()
		case <-r.connectionContext.Done():
			// this is caused when we lose connectivity
			return err
		}
	}
}

// Subscribe sends a "REQ" command to the relay r as in NIP-01. Events are
// returned through the channel sub.Events. The subscription is closed when
// context ctx is cancelled ("CLOSE" in NIP-01).
//
// Remember to cancel subscriptions, either by calling `.Unsub()` on them or
// ensuring their `context.Context` will be canceled at some point. Failure to
// do that will result in a huge number of halted goroutines being created.
func (r *Client) Subscribe(c cx, ff *filters.T,
	opts ...SubscriptionOption) (sub *Subscription, err er) {

	sub = r.PrepareSubscription(c, ff, opts...)
	if r.Connection == nil {
		err = errorf.E("not connected to %s", r.URL)
		return
	}
	if err = sub.Fire(); chk.T(err) {
		return nil, errorf.E("couldn't subscribe to %v at %s: %w",
			ff, r.URL, err)
	}
	return
}

// PrepareSubscription creates a subscription, but doesn't fire it.
//
// Remember to cancel subscriptions, either by calling `.Unsub()` on them or ensuring their `context.Context` will be canceled at some point.
// Failure to do that will result in a huge number of halted goroutines being created.
func (r *Client) PrepareSubscription(c cx, ff *filters.T,
	opts ...SubscriptionOption) *Subscription {

	current := subscriptionIDCounter.Add(1)
	c, cancel := context.Cancel(c)
	sub := &Subscription{
		Relay:             r,
		Context:           c,
		cancel:            cancel,
		counter:           no(current),
		Events:            make(event.C),
		EndOfStoredEvents: qu.Ts(1),
		ClosedReason:      make(chan st, 1),
		Filters:           ff,
	}
	for _, opt := range opts {
		switch o := opt.(type) {
		case WithLabel:
			sub.label = st(o)
		}
	}
	id := sub.GetID()
	r.Subscriptions.Store(id.String(), sub)
	// start handling events, eose, unsub etc:
	go sub.start()
	return sub
}

func (r *Client) QuerySync(ctx cx, f *filter.T,
	opts ...SubscriptionOption) ([]*event.T, er) {
	sub, err := r.Subscribe(ctx, filters.New(f), opts...)
	if err != nil {
		return nil, err
	}

	defer sub.Unsub()

	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		ctx, cancel = context.Timeout(ctx, 7*time.Second)
		defer cancel()
	}

	var events []*event.T
	for {
		select {
		case evt := <-sub.Events:
			if evt == nil {
				// channel is closed
				return events, nil
			}
			events = append(events, evt)
		case <-sub.EndOfStoredEvents:
			return events, nil
		case <-ctx.Done():
			return events, nil
		}
	}
}

func (r *Client) Count(c cx, ff *filters.T, opts ...SubscriptionOption) (no,
	er) {
	sub := r.PrepareSubscription(c, ff, opts...)
	sub.countResult = make(chan no)

	if err := sub.Fire(); chk.T(err) {
		return 0, err
	}

	defer sub.Unsub()

	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		c, cancel = context.Timeout(c, 7*time.Second)
		defer cancel()
	}

	for {
		select {
		case count := <-sub.countResult:
			return count, nil
		case <-c.Done():
			return 0, c.Err()
		}
	}
}

func (r *Client) Close() er {
	r.closeMutex.Lock()
	defer r.closeMutex.Unlock()
	if r.connectionContextCancel == nil {
		return errorf.E("relay already closed")
	}
	r.connectionContextCancel()
	r.connectionContextCancel = nil
	if r.Connection == nil {
		return errorf.E("relay not connected")
	}
	err := r.Connection.Close()
	r.Connection = nil
	if err != nil {
		return err
	}
	return nil
}
