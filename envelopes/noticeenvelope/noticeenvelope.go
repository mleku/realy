// Package noticeenvelope is a codec for the NOTICE envelope, which is used to
// serve (mostly ignored) messages that are supposed to be shown to a user in
// the client.
package noticeenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/text"
)

const L = "NOTICE"

type T struct {
	Message []byte
}

var _ codec.Envelope = (*T)(nil)

func New() *T                             { return &T{} }
func NewFrom[V string | []byte](msg V) *T { return &T{Message: []byte(msg)} }
func (en *T) Label() string               { return L }

func (en *T) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *T) Marshal(dst []byte) (b []byte) {
	var err error
	_ = err
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Message)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Message, r, err = text.UnmarshalQuoted(r); chk.E(err) {
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
