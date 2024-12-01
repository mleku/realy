package noticeenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/text"
)

const L = "NOTICE"

type T struct {
	Message by
}

var _ codec.Envelope = (*T)(nil)

func New() *T                     { return &T{} }
func NewFrom[V st | by](msg V) *T { return &T{Message: by(msg)} }
func (en *T) Label() string       { return L }

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
			o = append(o, '"')
			o = text.NostrEscape(o, en.Message)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) UnmarshalJSON(b by) (r by, err er) {
	r = b
	if en.Message, r, err = text.UnmarshalQuoted(r); chk.E(err) {
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
