// Package closeenvelope provides the encoder for the client message CLOSE which
// is a request to terminate a subscription.
package closeenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/subscription"
)

// L is the label associated with this type of codec.Envelope.
const L = "CLOSE"

// T is a CLOSE envelope, which is a signal from client to relay to stop a
// specified subscription.
type T struct {
	ID *subscription.Id
}

var _ codec.Envelope = (*T)(nil)

// New creates an empty new standard formatted closeenvelope.T.
func New() *T { return &T{ID: subscription.NewStd()} }

// NewFrom creates a new closeenvelope.T populated with subscription Id.
func NewFrom(id *subscription.Id) *T { return &T{ID: id} }

// Label returns the label of a closeenvelope.T.
func (en *T) Label() string { return L }

// Write the closeenvelope.T to a provided io.Writer.
func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal a closeenvelope.T envelope in minified JSON, appending to a provided
// destination slice.
func (en *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.ID.Marshal(o)
			return
		})
	return
}

// Unmarshal a closeenvelope.T from minified JSON, returning the remainder after
// the end of the envelope.
func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.ID, err = subscription.NewId([]byte{0}); chk.E(err) {
		return
	}
	if r, err = en.ID.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Parse reads a CLOSE envelope from minified JSON into a newly allocated
// closeenvelope.T.
func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
