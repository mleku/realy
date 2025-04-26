// Package closedenvelope defines the nostr message type CLOSED which is sent
// from a relay to indicate the relay-side termination of a subscription or the
// demand for authentication associated with a subscription.
package closedenvelope

import (
	"io"

	"realy.lol/chk"
	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/subscription"
	"realy.lol/text"
)

// L is the label associated with this type of codec.Envelope.
const L = "CLOSED"

// T is a CLOSED envelope, which is a signal that a subscription has been
// stopped on the relay side for some reason. Primarily this is for auth and can
// be for other things like rate limiting.
type T struct {
	Subscription *subscription.Id
	Reason       []byte
}

var _ codec.Envelope = (*T)(nil)

// New creates an empty new T.
func New() *T {
	return &T{Subscription: subscription.NewStd()}
}

// NewFrom creates a new closedenvelope.T populated with subscription Id and Reason.
func NewFrom(id *subscription.Id, msg []byte) *T { return &T{Subscription: id, Reason: msg} }

// Label returns the label of a closedenvelope.T.
func (en *T) Label() string { return L }

// ReasonString returns the Reason in the form of a string.
func (en *T) ReasonString() string { return string(en.Reason) }

// Write the closedenvelope.T to a provided io.Writer.
func (en *T) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

// Marshal a closedenvelope.T envelope in minified JSON, appending to a provided
// destination slice. Note that this ensures correct string escaping on the
// Reason field.
func (en *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.Subscription.Marshal(o)
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, en.Reason)
			o = append(o, '"')
			return
		})
	return
}

// Unmarshal a closedenvelope.T from minified JSON, returning the remainder after the end
// of the envelope. Note that this ensures the Reason string is correctly
// unescaped by NIP-01 escaping rules.
func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Subscription, err = subscription.NewId([]byte{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.Unmarshal(r); chk.E(err) {
		return
	}
	if en.Reason, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Parse reads a closedenvelope.T from minified JSON into a newly allocated closedenvelope.T.
func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
