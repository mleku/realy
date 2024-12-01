package eoseenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/subscription"
)

const L = "EOSE"

type T struct {
	Subscription *subscription.Id
}

var _ codec.Envelope = (*T)(nil)

func New() *T                        { return &T{Subscription: subscription.NewStd()} }
func NewFrom(id *subscription.Id) *T { return &T{Subscription: id} }
func (en *T) Label() string          { return L }

func (en *T) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *T) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst by) (o by, err er) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		},
	)
	return
}

func (en *T) UnmarshalJSON(b by) (r by, err er) {
	r = b
	if en.Subscription, err = subscription.NewId(by{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func Parse(b by) (t *T, rem by, err er) {
	t = New()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
