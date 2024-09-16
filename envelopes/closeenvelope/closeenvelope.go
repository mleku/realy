package closeenvelope

import (
	"io"

	"mleku.dev"
	"mleku.dev/envelopes"
	"mleku.dev/subscriptionid"
)

const L = "CLOSE"

type T struct {
	ID *subscriptionid.T
}

var _ realy.I = (*T)(nil)

func New() *T                         { return &T{ID: subscriptionid.NewStd()} }
func NewFrom(id *subscriptionid.T) *T { return &T{ID: id} }
func (en *T) Label() string           { return L }
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
	if en.ID, err = subscriptionid.New(B{0}); chk.E(err) {
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
