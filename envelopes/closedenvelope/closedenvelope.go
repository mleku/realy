package closedenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/subscription"
	"realy.lol/text"
)

const L = "CLOSED"

type T struct {
	Subscription *subscription.Id
	Reason       []byte
}

var _ codec.Envelope = (*T)(nil)

func New() *T                                    { return &T{Subscription: subscription.NewStd()} }
func NewFrom(id *subscription.Id, msg []byte) *T { return &T{Subscription: id, Reason: msg} }
func (en *T) Label() string                      { return L }
func (en *T) ReasonString() string               { return string(en.Reason) }

func (en *T) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

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

func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
