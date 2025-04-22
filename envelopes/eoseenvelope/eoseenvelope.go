// Package eoseenvelope provides an encoder for the EOSE (End Of Stored
// Events) event that signifies that a REQ has found all stored events and
// from here on the request morphs into a subscription, until the limit, if
// requested, or until CLOSE or CLOSED.
package eoseenvelope

import (
	"io"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/codec"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/subscription"
)

// L is the label associated with this type of codec.Envelope.
const L = "EOSE"

// T is an EOSE envelope (End of Stored Events), that signals the end of events
// that are stored and the beginning of a subscription. This is necessitated by
// the confusing multiplexing of websockets for multiple requests, and an ugly
// merging of two distinct API calls, filter and subscribe.
type T struct {
	Subscription *subscription.Id
}

var _ codec.Envelope = (*T)(nil)

// New creates a new eoseenvelope.T with a standard form subscription.Id.
func New() *T {
	return &T{Subscription: subscription.NewStd()}
}

// NewFrom creates a new  eoseenvelope.T using a provided subscription.Id.
func NewFrom(id *subscription.Id) *T { return &T{Subscription: id} }

// Label returns the label of a EOSE envelope.
func (en *T) Label() string { return L }

// Write the  eoseenvelope.T to a provided io.Writer.
func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal a eoseenvelope.T envelope in minified JSON, appending to a provided
// destination slice.
func (en *T) Marshal(dst []byte) (b []byte) {
	var err error
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.Subscription.Marshal(o)
			return
		},
	)
	_ = err
	return
}

// Unmarshal a eoseenvelope.T from minified JSON, returning the remainder after
// the end of the envelope.
func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Subscription, err = subscription.NewId([]byte{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Parse reads a EOSE envelope in minified JSON into a newly allocated
// eoseenvelope.T.
func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
