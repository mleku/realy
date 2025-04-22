// Package countenvelope is an encoder for the COUNT request (client) and
// response (relay) message types.
package countenvelope

import (
	"bytes"
	"io"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/codec"
	"realy.mleku.dev/envelopes"
	"realy.mleku.dev/errorf"
	"realy.mleku.dev/filters"
	"realy.mleku.dev/ints"
	"realy.mleku.dev/subscription"
	"realy.mleku.dev/text"
)

// L is the label associated with this type of codec.Envelope.
const L = "COUNT"

// Request is a COUNT envelope sent by a client to request a count of results.
// This is a stupid idea because it costs as much processing as fetching the
// events, but doesn't provide the means to actually get them (the HTTP API
// /filter does this by returning the actual event Ids).
type Request struct {
	Subscription *subscription.Id
	Filters      *filters.T
}

var _ codec.Envelope = (*Request)(nil)

// New creates a new Request with a standard style subscription.Id and empty filter.
func New() *Request {
	return &Request{Subscription: subscription.NewStd(),
		Filters: filters.New()}
}

// NewRequest creates a new Request with a provided subscription.Id and
// filter.T.
func NewRequest(id *subscription.Id, filters *filters.T) *Request {
	return &Request{Subscription: id,
		Filters: filters}
}

// Label returns the label of a CLOSED envelope.
func (en *Request) Label() string { return L }

// Write the Request to a provided io.Writer.
func (en *Request) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

// Marshal a Request appended to the provided destination slice as minified
// JSON.
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

// Unmarshal a Request from minified JSON, returning the remainder after the end
// of the envelope.
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

// ParseRequest reads a Request in minified JSON into a newly allocated Request.
func ParseRequest(b []byte) (t *Request, rem []byte, err error) {
	t = New()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}

// Response is a COUNT Response returning a count and approximate flag
// associated with the REQ subscription.Id.
type Response struct {
	ID          *subscription.Id
	Count       int
	Approximate bool
}

var _ codec.Envelope = (*Response)(nil)

// NewResponse creates a new empty countenvelope.Response with a standard formatted
// subscription.Id.
func NewResponse() *Response { return &Response{ID: subscription.NewStd()} }

// NewResponseFrom creates a new countenvelope.Response with provided string for the
// subscription.Id, a count and optional variadic approximate flag, which is
// otherwise false and does not get rendered into the JSON.
func NewResponseFrom[V string | []byte](s V, cnt int,
	approx ...bool) (res *Response, err error) {

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

// Label returns the COUNT label associated with a Response.
func (en *Response) Label() string { return L }

// Write a Response to a provided io.Writer as minified JSON.
func (en *Response) Write(w io.Writer) (err error) {
	_, err = w.Write(en.Marshal(nil))
	return
}

// Marshal a countenvelope.Response envelope in minified JSON, appending to a
// provided destination slice.
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

// Unmarshal a COUNT Response from minified JSON, returning the remainder after
// the end of the envelope.
func (en *Response) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	var inID, inCount bool
	for ; len(r) > 0; r = r[1:] {
		// first we should be finding a subscription Id
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

// Parse reads a Count Response in minified JSON into a newly allocated
// countenvelope.Response.
func Parse(b []byte) (t *Response, rem []byte, err error) {
	t = NewResponse()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
