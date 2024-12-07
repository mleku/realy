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
func (en *T) Write(w io.Writer) (err er) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *T) Marshal(dst by) (b by) {
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst by) (o by) {
			o = bst
			o = en.ID.Marshal(o)
			return
		})
	return
}

func (en *T) Unmarshal(b by) (r by, err er) {
	r = b
	if en.ID, err = subscription.NewId(by{0}); chk.E(err) {
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

func Parse(b by) (t *T, rem by, err er) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
