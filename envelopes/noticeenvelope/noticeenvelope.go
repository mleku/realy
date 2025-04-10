// Package noticeenvelope is a codec for the NOTICE envelope, which is used to
// serve (mostly ignored) messages that are supposed to be shown to a user in
// the client.
package noticeenvelope

import (
	"io"

	"realy.mleku.dev/codec"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/text"
)

// L is the label associated with this type of codec.Envelope.
const L = "NOTICE"

// T is a NOTICE envelope, intended to convey information to the user about the
// state of the relay connection. This thing is rarely displayed on clients
// except sometimes in event logs.
type T struct {
	Message []byte
}

var _ codec.Envelope = (*T)(nil)

// New creates a new empty NOTICE noticeenvelope.T.
func New() *T { return &T{} }

// NewFrom creates a new noticeenvelope.T with a provided message.
func NewFrom[V string | []byte](msg V) *T { return &T{Message: []byte(msg)} }

// Label returns the label of a NOTICE envelope.
func (en *T) Label() string { return L }

// Write the NOTICE T to a provided io.Writer.
func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal a NOTICE envelope in minified JSON into an noticeenvelope.T,
// appending to a provided destination slice. Note that this ensures correct
// string escaping on the Reason field.
func (en *T) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Message)
			o = append(o, '"')
			return
		})
	return
}

// Unmarshal a noticeenvelope.T from minified JSON, returning the remainder
// after the end of the envelope. Note that this ensures the Reason string is
// correctly unescaped by NIP-01 escaping rules.
func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Message, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

// Parse reads a NOTICE envelope in minified JSON into a newly allocated
// noticeenvelope.T.
func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
