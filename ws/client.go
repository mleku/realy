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

	"realy.mleku.dev/atomic"
	"realy.mleku.dev/auth"
	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/envelopes/closedenvelope"
	"realy.mleku.dev/envelopes/countenvelope"
	"realy.mleku.dev/envelopes/eoseenvelope"
	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/envelopes/noticeenvelope"
	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/errorf"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/log"
	"realy.mleku.dev/normalize"
	"realy.mleku.dev/signer"
)

var subscriptionIDCounter atomic.Int32

type Client struct {
	closeMutex                    sync.Mutex
	URL                           string
	RequestHeader                 http.Header // e.g. for origin header
	Connection                    *Connection
	Subscriptions                 *xsync.MapOf[string, *Subscription]
	ConnectionError               error
	connectionContext             context.T // will be canceled when the connection closes
	connectionContextCancel       context.F
	challenge                     []byte      // NIP-42 challenge, we only keep the last
	notices                       chan []byte // NIP-01 NOTICEs
	okCallbacks                   *xsync.MapOf[string, func(bool, string)]
	writeQueue                    chan writeRequest
	subscriptionChannelCloseQueue chan *Subscription
	signatureChecker              func(*event.T) bool
	AssumeValid                   bool // this will skip verifying signatures for events received from this relay
}

type writeRequest struct {
	msg    []byte
	answer chan error
}

// NewRelay returns a new relay. The relay connection will be closed when the context is canceled.
func NewRelay(c context.T, url string, opts ...RelayOption) *Client {
	ctx, cancel := context.Cancel(c)
	r := &Client{
		URL:                           string(normalize.URL([]byte(url))),
		connectionContext:             ctx,
		connectionContextCancel:       cancel,
		Subscriptions:                 xsync.NewMapOf[string, *Subscription](),
		okCallbacks:                   xsync.NewMapOf[string, func(bool, string)](),
		writeQueue:                    make(chan writeRequest),
		subscriptionChannelCloseQueue: make(chan *Subscription),
		signatureChecker:              func(e *event.T) bool { ok, _ := e.Verify(); return ok },
	}

	for _, opt := range opts {
		opt.ApplyRelayOption(r)
	}

	return r
}

// RelayConnect returns a relay object connected to url. Once successfully connected, cancelling
// ctx has no effect. To close the connection, call r.Close().
func RelayConnect(ctx context.T, url string, opts ...RelayOption) (*Client, error) {
	r := NewRelay(context.Bg(), url, opts...)
	err := r.Connect(ctx)
	return r, err
}

// RelayOption is the type of the argument passed for that.
type RelayOption interface {
	ApplyRelayOption(*Client)
}

var (
	_ RelayOption = (WithNoticeHandler)(nil)
	_ RelayOption = (WithSignatureChecker)(nil)
)

// WithNoticeHandler just takes notices and is expected to do something with them. when not
// given, defaults to logging the notices.
type WithNoticeHandler func(notice []byte)

func (nh WithNoticeHandler) ApplyRelayOption(r *Client) {
	r.notices = make(chan []byte)
	go func() {
		for notice := range r.notices {
			nh(notice)
		}
	}()
}

// WithSignatureChecker must be a function that checks the signature of an event and returns
// true or false.
type WithSignatureChecker func(*event.T) bool

func (sc WithSignatureChecker) ApplyRelayOption(r *Client) {
	r.signatureChecker = sc
}

// String just returns the relay URL.
func (r *Client) String() string {
	return r.URL
}

// Context retrieves the context that is associated with this relay connection.
func (r *Client) Context() context.T { return r.connectionContext }

// IsConnected returns true if the connection to this relay seems to be active.
func (r *Client) IsConnected() bool { return r.connectionContext.Err() == nil }

// Connect tries to establish a websocket connection to r.URL. If the context expires before the
// connection is complete, an error is returned. Once successfully connected, context expiration
// has no effect: call r.Close to close the connection.
//
// The underlying relay connection will use a background context. If you want to pass a custom
// context to the underlying relay connection, use NewRelay() and then Client.Connect().
func (r *Client) Connect(c context.T) error { return r.ConnectWithTLS(c, nil) }

// ConnectWithTLS tries to establish a secured websocket connection to r.URL using customized
// tls.Config (CA's, etc).
func (r *Client) ConnectWithTLS(ctx context.T, tlsConfig *tls.Config) error {
	if r.connectionContext == nil || r.Subscriptions == nil {
		return errorf.E("relay must be initialized with a call to NewRelay()")
	}
	if r.URL == "" {
		return errorf.E("invalid relay URL '%s'", r.URL)
	}
	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		ctx, cancel = context.Timeout(ctx, 7*time.Second)
		defer cancel()
	}
	conn, err := NewConnection(ctx, r.URL, r.RequestHeader, tlsConfig)
	if err != nil {
		return errorf.E("error opening websocket to '%s': %w", r.URL, err)
	}
	r.Connection = conn
	// ping every 29 seconds (??)
	ticker := time.NewTicker(29 * time.Second)
	// to be used when the connection is closed
	go func() {
		<-r.connectionContext.Done()
		// close these things when the connection is closed
		if r.notices != nil {
			close(r.notices)
		}
		// stop the ticker
		ticker.Stop()
		// close all subscriptions
		r.Subscriptions.Range(func(_ string, sub *Subscription) bool {
			go sub.Unsub()
			return true
		})
	}()
	// queue all write operations here so we don't do mutex spaghetti
	go WriteLoop(ticker, r)
	// general message reader loop
	go ReadLoop(conn, r)
	return nil
}

func ReadLoop(conn *Connection, r *Client) {
	var err error
	buf := new(bytes.Buffer)
	for {
		buf.Reset()
		if err = conn.ReadMessage(r.connectionContext, buf); chk.T(err) {
			r.ConnectionError = err
			r.Close()
			break
		}
		message := buf.Bytes()
		log.D.F("{%s} %v\n", r.URL, message)

		var t string
		if t, message = envelopes.Identify(message); chk.E(err) {
			continue
		}
		switch t {
		case noticeenvelope.L:
			env := noticeenvelope.New()
			if env, message, err = noticeenvelope.Parse(message); chk.E(err) {
				continue
			}
			// see WithNoticeHandler
			if r.notices != nil {
				r.notices <- env.Message
			} else {
				log.E.F("NOTICE from %s: '%s'\n", r.URL, env.Message)
			}
		case authenvelope.L:
			env := authenvelope.NewChallenge()
			if env, message, err = authenvelope.ParseChallenge(message); chk.E(err) {
				continue
			}
			if len(env.Challenge) == 0 {
				continue
			}
			r.challenge = env.Challenge
		case eventenvelope.L:
			env := eventenvelope.NewResult()
			if env, message, err = eventenvelope.ParseResult(message); chk.E(err) {
				continue
			}
			if len(env.Subscription.T) == 0 {
				continue
			}
			if sub, ok := r.Subscriptions.Load(env.Subscription.String()); !ok {
				log.D.F("{%s} no subscription with id '%s'\n", r.URL, env.Subscription)
				continue
			} else {
				// check if the event matches the desired filter, ignore otherwise
				if !sub.Filters.Match(env.Event) {
					log.D.F("{%s} filter does not match: %v ~ %v\n", r.URL,
						sub.Filters, env.Event)
					continue
				}
				// check signature, ignore invalid, except from trusted (AssumeValid) relays
				if !r.AssumeValid {
					if ok = r.signatureChecker(env.Event); !ok {
						log.E.F("{%s} bad signature on %s\n", r.URL, env.Event.Id)
						continue
					}
				}
				// dispatch this to the internal .events channel of the subscription
				sub.dispatchEvent(env.Event)
			}
		case eoseenvelope.L:
			env := eoseenvelope.New()
			if env, message, err = eoseenvelope.Parse(message); chk.E(err) {
				continue
			}
			if subscription, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
				subscription.dispatchEose()
			}
		case closedenvelope.L:
			env := closedenvelope.New()
			if env, message, err = closedenvelope.Parse(message); chk.E(err) {
				continue
			}
			if subscription, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
				subscription.dispatchClosed(env.ReasonString())
			}
		case countenvelope.L:
			env := countenvelope.NewResponse()
			if env, message, err = countenvelope.Parse(message); chk.E(err) {
				continue
			}
			if subscription, ok := r.Subscriptions.Load(env.ID.String()); ok && subscription.countResult != nil {
				subscription.countResult <- env.Count
			}
		case okenvelope.L:
			env := okenvelope.New()
			if env, message, err = okenvelope.Parse(message); chk.E(err) {
				continue
			}
			if okCallback, exist := r.okCallbacks.Load(env.EventID.String()); exist {
				okCallback(env.OK, env.ReasonString())
			} else {
				log.I.F("{%s} got an unexpected OK message for event %s", r.URL,
					env.EventID)
			}
		}
	}
}

func WriteLoop(ticker *time.Ticker, r *Client) {
	var err error
	for {
		select {
		case <-ticker.C:
			err = wsutil.WriteClientMessage(r.Connection.conn, ws.OpPing, nil)
			if err != nil {
				log.D.F("{%s} error writing ping: %v; closing websocket", r.URL,
					err)
				r.Close() // this should trigger a context cancelation
				return
			}
		case writeReq := <-r.writeQueue:
			// all write requests will go through this to prevent races
			if err = r.Connection.WriteMessage(r.connectionContext,
				writeReq.msg); chk.T(err) {
				writeReq.answer <- err
			}
			close(writeReq.answer)
		case <-r.connectionContext.Done():
			// stop here
			return
		}
	}
}

// Write queues a message to be sent to the relay.
func (r *Client) Write(msg []byte) <-chan error {
	ch := make(chan error)
	select {
	case r.writeQueue <- writeRequest{msg: msg, answer: ch}:
	case <-r.connectionContext.Done():
		go func() { ch <- errorf.E("connection closed") }()
	}
	return ch
}

// Publish sends an "EVENT" command to the relay r as in NIP-01 and waits for an OK response.
func (r *Client) Publish(c context.T, ev *event.T) error { return r.publish(c, ev) }

// Auth sends an "AUTH" command client->relay as in NIP-42 and waits for an OK response.
func (r *Client) Auth(c context.T, sign signer.I) error {
	authEvent := auth.CreateUnsigned(sign.Pub(), r.challenge, r.URL)
	if err := authEvent.Sign(sign); chk.T(err) {
		return errorf.E("error signing auth event: %w", err)
	}
	return r.publish(c, authEvent)
}

// publish can be used both for EVENT and for AUTH
func (r *Client) publish(ctx context.T, ev *event.T) (err error) {
	var cancel context.F
	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		ctx, cancel = context.TimeoutCause(ctx, 7*time.Second,
			errorf.E("given up waiting for an OK"))
		defer cancel()
	} else {
		// otherwise make the context cancellable so we can stop everything upon receiving an "OK"
		ctx, cancel = context.Cancel(ctx)
		defer cancel()
	}
	// listen for an OK callback
	gotOk := false
	id := ev.IdString()
	r.okCallbacks.Store(id, func(ok bool, reason string) {
		gotOk = true
		if !ok {
			err = errorf.E("msg: %s", reason)
		}
		cancel()
	})
	defer r.okCallbacks.Delete(id)
	// publish event
	var b []byte
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
			// this will be called when we get an OK or when the context has been canceled
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

// Subscribe sends a "REQ" command to the relay r as in NIP-01.
// Events are returned through the channel sub.Events.
// The subscription is closed when context ctx is cancelled ("CLOSE" in NIP-01).
//
// Remember to cancel subscriptions, either by calling `.Unsub()` on them or ensuring their `context.Context` will be canceled at some point.
// Failure to do that will result in a huge number of halted goroutines being created.
func (r *Client) Subscribe(c context.T, ff *filters.T,
	opts ...SubscriptionOption) (*Subscription, error) {
	sub := r.PrepareSubscription(c, ff, opts...)
	if r.Connection == nil {
		return nil, errorf.E("not connected to %s", r.URL)
	}
	if err := sub.Fire(); chk.T(err) {
		return nil, errorf.E("couldn't subscribe to %v at %s: %w", ff, r.URL, err)
	}
	return sub, nil
}

// PrepareSubscription creates a subscription, but doesn't fire it.
//
// Remember to cancel subscriptions, either by calling `.Unsub()` on them or ensuring their `context.Context` will be canceled at some point.
// Failure to do that will result in a huge number of halted goroutines being created.
func (r *Client) PrepareSubscription(c context.T, ff *filters.T,
	opts ...SubscriptionOption) *Subscription {
	current := subscriptionIDCounter.Add(1)
	c, cancel := context.Cancel(c)
	sub := &Subscription{
		Relay:             r,
		Context:           c,
		cancel:            cancel,
		counter:           int(current),
		Events:            make(event.C),
		EndOfStoredEvents: make(chan struct{}, 1),
		ClosedReason:      make(chan string, 1),
		Filters:           ff,
	}
	for _, opt := range opts {
		switch o := opt.(type) {
		case WithLabel:
			sub.label = string(o)
		}
	}
	id := sub.GetID()
	r.Subscriptions.Store(id.String(), sub)
	// start handling events, eose, unsub etc:
	go sub.start()
	return sub
}

// QuerySync is only used in tests. The realy query method is synchronous now anyway (it ensures
// sort order is respected).
func (r *Client) QuerySync(ctx context.T, f *filter.T,
	opts ...SubscriptionOption) ([]*event.T, error) {
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

// TODO: count is a dumb idea anyway, and nothing is using this
// func (r *Client) Count(c context.T, ff *filters.T, opts ...SubscriptionOption) (int, error) {
// 	sub := r.PrepareSubscription(c, ff, opts...)
// 	sub.countResult = make(chan int)
//
// 	if err := sub.Fire(); chk.T(err) {
// 		return 0, err
// 	}
//
// 	defer sub.Unsub()
//
// 	if _, ok := c.Deadline(); !ok {
// 		// if no timeout is set, force it to 7 seconds
// 		var cancel context.F
// 		c, cancel = context.Timeout(c, 7*time.Second)
// 		defer cancel()
// 	}
//
// 	for {
// 		select {
// 		case count := <-sub.countResult:
// 			return count, nil
// 		case <-c.Done():
// 			return 0, c.Err()
// 		}
// 	}
// }

// Close shuts down a websocket client connection.
func (r *Client) Close() error {
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
