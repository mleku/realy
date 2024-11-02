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
	ID      *subscription.Id
	Filters *filters.T
}

var _ codec.Envelope = (*Request)(nil)

func New() *Request {
	return &Request{ID: subscription.NewStd(),
		Filters: filters.New()}
}
func NewRequest(id *subscription.Id, filters *filters.T) *Request {
	return &Request{ID: id,
		Filters: filters}
}
func (en *Request) Label() string { return L }
func (en *Request) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Request) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.ID.MarshalJSON(o); chk.E(err) {
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

func (en *Request) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.ID, err = subscription.NewId(B{0}); chk.E(err) {
		return
	}
	if r, err = en.ID.UnmarshalJSON(r); chk.E(err) {
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

func ParseRequest(b B) (t *Request, rem B, err E) {
	t = New()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
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
func NewResponseFrom[V S | B](s V, cnt N, approx ...bool) (res *Response, err E) {
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
func (en *Response) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Response) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
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

func (en *Response) UnmarshalJSON(b B) (r B, err error) {
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
				if r, err = n.UnmarshalJSON(r); chk.E(err) {
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
							if bytes.Contains(r[:i], B("true")) {
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

func Parse(b B) (t *Response, rem B, err E) {
	t = NewResponse()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
