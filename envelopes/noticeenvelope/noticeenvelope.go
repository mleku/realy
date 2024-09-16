package noticeenvelope

import (
	"io"

	"mleku.dev"
	"mleku.dev/envelopes"
	"mleku.dev/text"
)

const L = "NOTICE"

type T struct {
	Message B
}

var _ realy.I = (*T)(nil)

func New() *T                   { return &T{} }
func NewFrom[V S | B](msg V) *T { return &T{Message: B(msg)} }
func (en *T) Label() string     { return L }

func (en *T) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *T) MarshalJSON(dst B) (b B, err E) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Message)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) UnmarshalJSON(b B) (r B, err E) {
	r = b
	if en.Message, r, err = text.UnmarshalQuoted(r); chk.E(err) {
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
