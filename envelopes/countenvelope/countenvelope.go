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
func (en *Request) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Request) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst by) (o by, err er) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = en.Filters.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Request) UnmarshalJSON(b by) (r by, err er) {
	r = b
	if en.Subscription, err = subscription.NewId(by{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	en.Filters = filters.New()
	if r, err = en.Filters.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseRequest(b by) (t *Request, rem by, err er) {
	t = New()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}

type Response struct {
	ID          *subscription.Id
	Count       no
	Approximate bo
}

var _ codec.Envelope = (*Response)(nil)

func NewResponse() *Response { return &Response{ID: subscription.NewStd()} }
func NewResponseFrom[V st | by](s V, cnt no, approx ...bo) (res *Response, err er) {
	var a bo
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
func (en *Response) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Response) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst by) (o by, err er) {
			o = bst
			if o, err = en.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			c := ints.New(en.Count)
			o, err = c.MarshalJSON(o)
			if en.Approximate {
				o = append(dst, ',')
				o = append(o, "true"...)
			}
			return
		})
	return
}

func (en *Response) UnmarshalJSON(b by) (r by, err er) {
	r = b
	var inID, inCount bo
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
				if r, err = n.UnmarshalJSON(r); chk.E(err) {
					return
				}
				en.Count = no(n.Uint64())
			} else {
				// can only be either the end or optional approx
				if r[0] == ']' {
					return
				} else {
					for i := range r {
						if r[i] == ']' {
							if bytes.Contains(r[:i], by("true")) {
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

func Parse(b by) (t *Response, rem by, err er) {
	t = NewResponse()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
