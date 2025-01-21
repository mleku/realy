// Package ws is a websocket library
//
// todo: this client code is bullshit
package ws

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/puzpuzpuz/xsync/v2"

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
	"realy.lol/normalize"
	"realy.lol/eventid"
	"realy.lol/relayinfo"
	"fmt"
	"realy.lol/qu"
	"realy.lol/signer"
	"realy.lol/p256k/sign"
	"realy.lol/codec"
)

type Status int

var subscriptionIDCounter atomic.Int32

type Client struct {
	// Ctx will be canceled when connection closes
	Ctx                     context.T
	ConnectionContextCancel context.F
	closeMutex              sync.Mutex
	url                     by
	// RequestHeader  e.g. for origin header
	RequestHeader   http.Header
	Connection      *Connection
	Subscriptions   *xsync.MapOf[st, *Subscription]
	ConnectionError er
	done            sync.Once
	// challenge is NIP-42 challenge, only keep the last
	challenge    by
	AuthRequired qu.C
	AuthEventID  *eventid.T
	Authed       qu.C
	// notices are NIP-01 NOTICE
	notices                       chan by
	okCallbacks                   *xsync.MapOf[st, func(bo, by)]
	writeQueue                    chan writeRequest
	subscriptionChannelCloseQueue chan *Subscription

	// custom things that aren't often used
	//
	AssumeValid bool // skip verifying signatures of events from this relay
}

func (r *Client) URL() st { return st(r.url) }

func (r *Client) Delete(key string) { r.Subscriptions.Delete(key) }

type writeRequest struct {
	msg    []byte
	answer chan error
}

// NewClient returns a new relay client. The relay connection will be closed when the
// context is canceled.
func NewClient(c context.T, url string, opts ...Option) *Client {
	ctx, cancel := context.Cancel(c)
	r := &Client{
		url:                           normalize.URL(url),
		Ctx:                           ctx,
		ConnectionContextCancel:       cancel,
		Subscriptions:                 xsync.NewMapOf[*Subscription](),
		okCallbacks:                   xsync.NewMapOf[func(bo, by)](),
		writeQueue:                    make(chan writeRequest),
		subscriptionChannelCloseQueue: make(chan *Subscription),
		AuthRequired:                  make(chan struct{}),
		Authed:                        make(chan struct{}),
	}

	for _, opt := range opts {
		switch o := opt.(type) {
		case WithNoticeHandler:
			r.notices = make(chan by)
			go func() {
				for n := range r.notices {
					o(n)
				}
			}()
		}
	}

	return r
}

// Connect returns a relay object connected to url. Once successfully
// connected, cancelling ctx has no effect. To close the connection, call
// r.Close().
func Connect(c context.T, url string, opts ...Option) (*Client, error) {
	r := NewClient(c, url, opts...)
	err := r.Connect(c)
	return r, err
}

// ConnectWithAuth auths with the relay, checks if its NIP-11 says auth-required
// and uses the provided sec to sign the auth challenge.
func ConnectWithAuth(c context.T, url st, sign signer.I,
	opts ...Option) (rl *Client, err error) {

	if rl, err = Connect(c, url, opts...); chk.E(err) {
		return
	}
	var inf *relayinfo.T
	if inf, err = relayinfo.Fetch(c, url); chk.E(err) {
		return
	}
	// if NIP-11 doesn't say auth-required, we are done
	if !inf.Limitation.AuthRequired {
		return
	}
	// otherwise, expect auth immediately and sign on it. some relays may not send
	// the auth challenge without being prompted by a req envelope but fuck them.
	// auth-required in nip-11 should mean auth on connect. period.
	authed := false
out:
	for i := 0; i < 2; i++ {
		// but just in case, we will do this twice if need be. The first try may
		// time out because the relay waits for a req, or because the auth
		// doesn't trigger until a message is received.
		select {
		case <-rl.AuthRequired:
			if err = rl.Auth(c, sign); chk.E(err) {
				return
			}
		case <-time.After(5 * time.Second):
		case <-rl.Authed:
			log.T.Ln("authed to relay", rl.AuthEventID)
			authed = true
		}
		if authed {
			break out
		}
		// to trigger this if auth wasn't immediately demanded, send out a dummy
		// empty req.
		filt := filters.New(&filter.T{Limit: filter.L(1)})
		var sub *Subscription
		if sub, err = rl.Subscribe(c, filt); chk.E(err) {
			// not sure what to do here
		}
		sub.Close()
		// at this point if we haven't received an auth there is something wrong
		// with the relay.
	}
	return
}

// When instantiating relay connections, some options may be passed.

// Option is the type of the argument passed for that.
type Option interface {
	IsRelayOption()
}

// WithNoticeHandler just takes notices and is expected to do something with
// them. when not given, defaults to logging the notices.
type WithNoticeHandler func(notice by)

func (_ WithNoticeHandler) IsRelayOption() {}

var _ Option = (WithNoticeHandler)(nil)

// String just returns the relay URL.
func (r *Client) String() st {
	return st(r.url)
}

// Context retrieves the context that is associated with this relay connection.
func (r *Client) Context() context.T { return r.Ctx }

// IsConnected returns true if the connection to this relay seems to be active.
func (r *Client) IsConnected() bool { return r.Ctx.Err() == nil }

// Connect tries to establish a websocket connection to r.URL. If the context
// expires before the connection is complete, an error is returned. Once
// successfully connected, context expiration has no effect: call r.Close to
// close the connection.
//
// The underlying relay connection will use a background context. If you want to
// pass a custom context to the underlying relay connection, use NewClient() and
// then Relay.Connect().
func (r *Client) Connect(c context.T) (err error) {
	if r.Ctx == nil || r.Subscriptions == nil {
		return fmt.Errorf("relay must be initialized with a call to NewClient()")
	}
	if len(r.url) < 1 {
		return fmt.Errorf("invalid relay URL '%s'", r.URL())
	}
	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		c, cancel = context.Timeout(c, 7*time.Second)
		defer cancel()
	}
	var conn *Connection
	conn, err = NewConnection(c, r.url, r.RequestHeader, nil)
	if err != nil {
		return fmt.Errorf("error opening websocket to '%s': %w", r.URL(), err)
	}
	r.Connection = conn
	// ping every 29 seconds
	ticker := time.NewTicker(29 * time.Second)
	// to be used when the connection is closed
	go func() {
		<-r.Ctx.Done()
		// close these things when the connection is closed
		if r.notices != nil {
			log.I.Ln("closing notices chan")
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
	go func() {
		var err error
		for {
			select {
			case <-ticker.C:
				err = wsutil.WriteClientMessage(r.Connection.Conn, ws.OpPing,
					nil)
				if err != nil {
					log.D.F("{%s} error writing ping: %v; closing websocket",
						r.URL(), err)
					chk.D(r.Close()) // this should trigger a context cancelation
					return
				}
			case wr := <-r.writeQueue:
				if wr.msg == nil {
					return
				}
				// all write requests will go through this to prevent races
				if err = r.Connection.WriteMessage(r.Ctx,
					wr.msg); err != nil {
					wr.answer <- err
				}
				close(wr.answer)
			case <-r.Ctx.Done():
				// stop here
				return
			}
		}
	}()

	// general message reader loop
	go r.MessageReadLoop(conn)
	return nil
}

func (r *Client) MessageReadLoop(conn *Connection) {
	var err error
	for {
		buf := new(bytes.Buffer)
		if err = conn.ReadMessage(r.Ctx, buf); err != nil {
			r.ConnectionError = err
			chk.D(r.Close())
			break
		}

		message := buf.Bytes()
		// log.I.F("{%s} received %v", r.URL(), string(message))
		var rem by
		var t st
		if t, rem, err = envelopes.Identify(message); chk.E(err) {
			log.I.Ln(string(message))
			continue
		}
		if t == "" {
			continue
		}

		switch t {
		case noticeenvelope.L:
			env := noticeenvelope.New()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			// see WithNoticeHandler
			if r.notices != nil {
				r.notices <- env.Message
			} else {
				log.D.F("NOTICE from %s: '%s'", r.URL(), env.Message)
			}

		case authenvelope.L:
			env := authenvelope.NewChallenge()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			r.challenge = env.Challenge
			log.D.F("received challenge %s", r.challenge)
			r.AuthRequired <- struct{}{}

		case eventenvelope.L:
			env := eventenvelope.NewResult()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			// if it has no subscription ID we don't know what it is
			if env.Subscription.String() == "" {
				continue
			}
			if s, ok := r.Subscriptions.Load(env.Subscription.String()); !ok {
				log.D.F("{%s} no subscription with id '%s'",
					r.URL(), env.Subscription.String())
				continue
			} else {
				// check if the event matches the desired filter, ignore otherwise
				if !s.Filters.Match(env.Event) {
					log.D.F("{%s} filter does not match: %s ~ %s",
						r.URL(), s.Filters.String(), env.Event.Serialize())
					continue
				}
				// check signature, ignore invalid, except from trusted (AssumeValid) relays
				if !r.AssumeValid {
					if ok, err = env.Event.CheckSignature(); !ok {
						errmsg := ""
						if chk.D(err) {
							errmsg = err.Error()
						}
						log.D.F("{%s} bad signature on %s; %s",
							r.URL(), env.Event.ID, errmsg)
						continue
					}
				}
				// dispatch this to the internal .events channel of the
				// subscription
				s.dispatchEvent(env.Event)
			}

		case eoseenvelope.L:
			env := eoseenvelope.New()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			log.D.Ln("eose", r.Subscriptions.Size())
			if s, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
				log.D.Ln("dispatching eose", env.Subscription.String())
				s.dispatchEose()
			}

		case closedenvelope.L:
			env := closedenvelope.New()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			if s, ok := r.Subscriptions.Load(env.Subscription.String()); ok {
				s.dispatchClosed(env.Reason)
			}

		case countenvelope.L:
			env := countenvelope.NewResponse()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			if s, ok := r.Subscriptions.Load(env.ID.String()); ok &&
				s.countResult != nil {
				s.countResult <- env.Count
			}

		case okenvelope.L:
			env := okenvelope.New()
			if rem, err = env.Unmarshal(rem); chk.E(err) {
				continue
			}
			if env.EventID == r.AuthEventID {
				close(r.Authed)
			}
			if okCallback, exist := r.okCallbacks.Load(env.EventID.String()); exist {
				okCallback(env.OK, env.Reason)
			} else {
				log.D.F("{%s} got an unexpected OK message for event %s",
					r.URL(), env.EventID)
			}
		}
	}
}

// Write queues a message to be sent to the relay.
func (r *Client) Write(msg []byte) (ch chan error) {
	ch = make(chan error)
	timeout := time.After(time.Second * 5)
	select {
	case r.writeQueue <- writeRequest{msg: msg, answer: ch}:
	case <-r.Ctx.Done():
		ch <- fmt.Errorf("connection closed")
	case <-timeout:
		ch <- fmt.Errorf("write timed out")
		return
	}
	return
}

// Publish sends an "EVENT" command to the relay r as in NIP-01 and waits for an
// OK response.
func (r *Client) Publish(c context.T, ev *event.T) error {
	return r.publish(c, st(ev.ID), eventenvelope.NewSubmissionWith(ev))
}

// Auth sends an "AUTH" command client->relay as in NIP-42 and waits for an OK
// response.
func (r *Client) Auth(c context.T, s signer.I) (err er) {
	log.I.Ln("sending auth response to relay", r.URL())
	authEvent := auth.CreateUnsigned(r.challenge, r.URL())
	if authEvent, err = sign.SignEvent(s, authEvent); chk.D(err) {
		return fmt.Errorf("error signing auth event: %w", err)
	}
	return r.publish(c, st(authEvent.ID),
		&authenvelope.Response{Event: authEvent})
}

// publish can be used both for EVENT and for AUTH
func (r *Client) publish(c context.T, id st, env codec.Envelope) (err error) {
	var cancel context.F
	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 4 seconds
		c, cancel = context.Timeout(c, 4*time.Second)
		defer cancel()
	} else {
		// otherwise make the context cancellable, so we can stop everything
		// upon receiving an "OK"
		c, cancel = context.Cancel(c)
		defer cancel()
	}
	// listen for an OK callback
	gotOk := false
	r.okCallbacks.Store(id, func(ok bo, reason by) {
		gotOk = true
		if !ok {
			err = log.E.Err("msg: %s", reason)
		}
		cancel()
	})
	defer r.okCallbacks.Delete(id)
	// publish event
	var enb []byte
	enb = env.Marshal(enb)
	// log.T.F("{%s} sending %v", r.URL(), string(enb))
	if err = <-r.Write(enb); err != nil {
		return err
	}
	for {
		select {
		case <-c.Done():
			// this will be called when we get an OK or when the context has
			// been canceled
			if gotOk {
				return err
			}
			return c.Err()
		case <-r.Ctx.Done():
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
// ensuring their `context.T` will be canceled at some point. Failure to do that
// will result in a huge number of halted goroutines being created.
func (r *Client) Subscribe(c context.T, f *filters.T,
	opts ...SubscriptionOption) (sub *Subscription, err er) {

	sub = r.PrepareSubscription(c, f, opts...)

	if err := sub.Fire(); err != nil {
		return nil, fmt.Errorf("couldn't subscribe to %v at %s: %w", f, r.URL(),
			err)
	}

	return sub, nil
}

// PrepareSubscription creates a subscription, but doesn't fire it.
//
// Remember to cancel subscriptions, either by calling `.Unsub()` on them or
// ensuring their `context.T` will be canceled at some point. Failure to do that
// will result in a huge number of halted goroutines being created.
func (r *Client) PrepareSubscription(c context.T, f *filters.T,
	opts ...SubscriptionOption) *Subscription {

	if r.Connection == nil {
		panic(fmt.Errorf("must call .Connect() first before calling .Subscribe()"))
	}

	current := subscriptionIDCounter.Add(1)
	ctx, cancel := context.Cancel(c)

	sub := &Subscription{
		Relay:             r,
		Context:           ctx,
		cancel:            cancel,
		counter:           int(current),
		Events:            make(event.C),
		EndOfStoredEvents: make(chan struct{}),
		ClosedReason:      make(chan by, 1),
		Filters:           f,
	}

	for _, opt := range opts {
		switch o := opt.(type) {
		case WithLabel:
			sub.label = o
		}
	}

	id := sub.GetID()
	r.Subscriptions.Store(id.String(), sub)

	// start handling events, eose, unsub etc:
	go sub.start()

	return sub
}

func (r *Client) QuerySync(c context.T, f *filter.T,
	opts ...SubscriptionOption) ([]*event.T, error) {
	log.D.F("%s", f.Serialize())
	sub, err := r.Subscribe(c, filters.New(f), opts...)
	if err != nil {
		return nil, err
	}

	defer sub.Unsub()

	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		c, cancel = context.Timeout(c, 7*time.Second)
		defer cancel()
	}

	var events []*event.T
	for {
		select {
		case evt := <-sub.Events:
			if evt == nil {
				log.I.Ln("channel is closed")
				return events, nil
			}
			events = append(events, evt)
		case <-sub.EndOfStoredEvents:
			log.I.Ln("EOSE")
			return events, nil
		case <-c.Done():
			log.I.Ln("sub context done")
			return events, nil
		}
	}
}

func (r *Client) Count(c context.T, ff *filters.T,
	opts ...SubscriptionOption) (int, error) {

	sub := r.PrepareSubscription(c, ff, opts...)
	sub.countResult = make(chan int)

	if err := sub.Fire(); chk.E(err) {
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

func (r *Client) Close() error {
	r.closeMutex.Lock()
	defer r.closeMutex.Unlock()

	if r.ConnectionContextCancel == nil {
		return fmt.Errorf("relay not connected")
	}

	r.ConnectionContextCancel()
	r.ConnectionContextCancel = nil
	return r.Connection.Conn.Close()
}
