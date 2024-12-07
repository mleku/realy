package reqenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/filters"
	"realy.lol/subscription"
	"realy.lol/text"
)

const L = "REQ"

type T struct {
	Subscription *subscription.Id
	Filters      *filters.T
}

var _ codec.Envelope = (*T)(nil)

func New() *T {
	return &T{Subscription: subscription.NewStd(),
		Filters: filters.New()}
}
func NewFrom(id *subscription.Id, filters *filters.T) *T {
	return &T{Subscription: id,
		Filters: filters}
}
func (en *T) Label() string { return L }

func (en *T) Write(w io.Writer) (err er) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *T) Marshal(dst by) (b by) {
	var err er
	_ = err
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst by) (o by) {
			o = bst
			o = en.Subscription.Marshal(o)
			for _, f := range en.Filters.F {
				o = append(o, ',')
				o = f.Marshal(o)
			}
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
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	en.Filters = filters.New()
	if r, err = en.Filters.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func (en *T) Parse(b by) (t *T, rem by, err er) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
