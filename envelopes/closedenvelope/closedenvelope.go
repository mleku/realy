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
	Reason       by
}

var _ codec.Envelope = (*T)(nil)

func New() *T                                { return &T{Subscription: subscription.NewStd()} }
func NewFrom(id *subscription.Id, msg by) *T { return &T{Subscription: id, Reason: msg} }
func (en *T) Label() string                  { return L }
func (en *T) ReasonString() string           { return st(en.Reason) }

func (en *T) Write(w io.Writer) (err er) {
	var b by
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

func (en *T) Marshal(dst by) (b by) {
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst by) (o by) {
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

func (en *T) Unmarshal(b by) (r by, err er) {
	r = b
	if en.Subscription, err = subscription.NewId(by{0}); chk.E(err) {
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

func Parse(b by) (t *T, rem by, err er) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
