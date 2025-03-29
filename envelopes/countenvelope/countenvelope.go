// Package countenvelope is an encoder for the COUNT request (client) and
// response (relay) message types.
package countenvelope

import (
	"bytes"
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/filters"
	"realy.lol/ints"
	"realy.lol/subscription"
	"realy.lol/text"
)

const L = "COUNT"

type Request struct {
	Subscription *subscription.Id
	Filters      *filters.T
}

var _ codec.Envelope = (*Request)(nil)

func New() *Request {
	return &Request{Subscription: subscription.NewStd(),
		Filters: filters.New()}
}
func NewRequest(id *subscription.Id, filters *filters.T) *Request {
	return &Request{Subscription: id,
		Filters: filters}
}
func (en *Request) Label() string { return L }
func (en *Request) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

func (en *Request) Marshal(dst []byte) (b []byte) {
	var err error
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.Subscription.Marshal(o)
			o = append(o, ',')
			o = en.Filters.Marshal(o)
			return
		})
	_ = err
	return
}

func (en *Request) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	if en.Subscription, err = subscription.NewId([]byte{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.Unmarshal(r); chk.E(err) {
		return
	}
	en.Filters = filters.New()
	if r, err = en.Filters.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseRequest(b []byte) (t *Request, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}

type Response struct {
	ID          *subscription.Id
	Count       int
	Approximate bool
}

var _ codec.Envelope = (*Response)(nil)

func NewResponse() *Response { return &Response{ID: subscription.NewStd()} }
func NewResponseFrom[V string | []byte](s V, cnt int, approx ...bool) (res *Response, err error) {
	var a bool
	if len(approx) > 0 {
		a = approx[0]
	}
	if len(s) < 0 || len(s) > 64 {
		err = errorf.E("subscription id must be length > 0 and <= 64")
		return
	}
	return &Response{subscription.MustNew(s), cnt, a}, nil
}
func (en *Response) Label() string { return L }
func (en *Response) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *Response) Marshal(dst []byte) (b []byte) {
	var err error
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = en.ID.Marshal(o)
			o = append(o, ',')
			c := ints.New(en.Count)
			o = c.Marshal(o)
			if en.Approximate {
				o = append(dst, ',')
				o = append(o, "true"...)
			}
			return
		})
	_ = err
	return
}

func (en *Response) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	var inID, inCount bool
	for ; len(r) > 0; r = r[1:] {
		// first we should be finding a subscription ID
		if !inID && r[0] == '"' {
			r = r[1:]
			// so we don't do this twice
			inID = true
			for i := range r {
				if r[i] == '\\' {
					continue
				} else if r[i] == '"' {
					// skip escaped quotes
					if i > 0 {
						if r[i-1] != '\\' {
							continue
						}
					}
					if en.ID, err = subscription.
						NewId(text.NostrUnescape(r[:i])); chk.E(err) {

						return
					}
					// trim the rest
					r = r[i:]
				}
			}
		} else {
			// pass the comma
			if r[0] == ',' {
				continue
			} else if !inCount {
				inCount = true
				n := ints.New(0)
				if r, err = n.Unmarshal(r); chk.E(err) {
					return
				}
				en.Count = int(n.Uint64())
			} else {
				// can only be either the end or optional approx
				if r[0] == ']' {
					return
				} else {
					for i := range r {
						if r[i] == ']' {
							if bytes.Contains(r[:i], []byte("true")) {
								en.Approximate = true
							}
							return
						}
					}
				}
			}
		}
	}
	return
}

func Parse(b []byte) (t *Response, rem []byte, err error) {
	t = NewResponse()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
