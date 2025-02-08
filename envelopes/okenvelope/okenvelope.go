package okenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/eventid"
	"realy.lol/sha256"
	"realy.lol/text"
)

const (
	L = "OK"
)

type T struct {
	EventID *eventid.T
	OK      bool
	Reason  []byte
}

var _ codec.Envelope = (*T)(nil)

func New() *T { return &T{} }
func NewFrom[V string | []byte](eid V, ok bool, msg ...[]byte) *T {
	var m []byte
	if len(msg) > 0 {
		m = msg[0]
	}
	if len(eid) != sha256.Size {
		log.W.F("event ID unexpected length, expect %d got %d",
			len(eid), sha256.Size)
	}
	return &T{EventID: eventid.NewWith(eid), OK: ok, Reason: m}
}
func (en *T) Label() string        { return L }
func (en *T) ReasonString() string { return string(en.Reason) }

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
			o = en.EventID.ByteString(o)
			o = append(o, '"')
			o = append(o, ',')
			o = text.MarshalBool(o, en.OK)
			o = append(o, ',')
			o = append(o, '"')
			o = text.NostrEscape(o, en.Reason)
			o = append(o, '"')
			return
		})
	return
}

func (en *T) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	var idHex []byte
	if idHex, r, err = text.UnmarshalHex(r); chk.E(err) {
		return
	}
	if len(idHex) != sha256.Size {
		err = errorf.E("invalid size for ID, require %d got %d",
			len(idHex), sha256.Size)
	}
	en.EventID = eventid.NewWith(idHex)
	if r, err = text.Comma(r); chk.E(err) {
		return
	}
	if r, en.OK, err = text.UnmarshalBool(r); chk.E(err) {
		return
	}
	if r, err = text.Comma(r); chk.E(err) {
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

func Parse(b []byte) (t *T, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
