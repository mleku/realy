package closedenvelope

import (
	"io"

	"mleku.dev"
	"mleku.dev/envelopes"
	"mleku.dev/subscriptionid"
	"mleku.dev/text"
)

const L = "CLOSED"

type T struct {
	Subscription *subscriptionid.T
	Reason       B
}

var _ realy.I = (*T)(nil)

func New() *T                                { return &T{Subscription: subscriptionid.NewStd()} }
func NewFrom(id *subscriptionid.T, msg B) *T { return &T{Subscription: id, Reason: msg} }
func (en *T) Label() string                  { return L }
func (en *T) ReasonString() string           { return S(en.Reason) }

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
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, en.Reason)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.Subscription, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
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

func Parse(b B) (t *T, rem B, err E) {
	t = New()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
