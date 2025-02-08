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
func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *T) Marshal(dst []byte) (b []byte) {
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.ID.Marshal(o)
			return
		})
	return
}

func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.ID, err = subscription.NewId([]byte{0}); chk.E(err) {
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

func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
