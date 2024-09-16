package reqenvelope

import (
	"io"

	"realy.lol"
	"realy.lol/envelopes"
	"realy.lol/filters"
	sid "realy.lol/subscriptionid"
	"realy.lol/text"
)

const L = "REQ"

type T struct {
	Subscription *sid.T
	Filters      *filters.T
}

var _ realy.I = (*T)(nil)

func New() *T {
	return &T{Subscription: sid.NewStd(),
		Filters: filters.New()}
}
func NewFrom(id *sid.T, filters *filters.T) *T { return &T{Subscription: id, Filters: filters} }
func (en *T) Label() string                    { return L }

func (en *T) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			for _, f := range en.Filters.F {
				o = append(o, ',')
				if o, err = f.MarshalJSON(o); chk.E(err) {
					return
				}
			}
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.Subscription, err = sid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	en.Filters = filters.New()
	if r, err = en.Filters.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func (en *T) Parse(b B) (t *T, rem B, err E) {
	t = New()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
