package closeenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/subscription"
)

const L = "CLOSE"

type T struct {
	ID *subscription.Id
}

var _ codec.Envelope = (*T)(nil)

func New() *T                        { return &T{ID: subscription.NewStd()} }
func NewFrom(id *subscription.Id) *T { return &T{ID: id} }
func (en *T) Label() string          { return L }
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
			if o, err = en.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.ID, err = subscription.NewId(B{0}); chk.E(err) {
		return
	}
	if r, err = en.ID.UnmarshalJSON(r); chk.E(err) {
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
