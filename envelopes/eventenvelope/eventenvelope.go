package eventenvelope

import (
	"io"

	"realy.lol/codec"
	"realy.lol/envelopes"
	"realy.lol/event"
	"realy.lol/subscription"
)

const L = "EVENT"

// Submission is a request from a client for a realy to store an event.
type Submission struct {
	*event.T
}

var _ codec.Envelope = (*Submission)(nil)

func NewSubmission() *Submission                { return &Submission{T: &event.T{}} }
func NewSubmissionWith(ev *event.T) *Submission { return &Submission{T: ev} }
func (en *Submission) Label() string            { return L }

func (en *Submission) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Submission) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.T.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Submission) UnmarshalJSON(b B) (r B, err error) {
	r = b
	en.T = event.New()
	if r, err = en.T.UnmarshalJSON(r); chk.T(err) {
		return
	}
	if r, err = en.T.MarshalJSON(nil); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseSubmission(b B) (t *Submission, rem B, err E) {
	t = NewSubmission()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}

// Result is an event matching a filter associated with a subscription.
type Result struct {
	Subscription *subscription.Id
	Event        *event.T
}

var _ codec.Envelope = (*Result)(nil)

func NewResult() *Result { return &Result{} }
func NewResultWith[V S | B](s V, ev *event.T) (res *Result, err E) {
	if len(s) < 0 || len(s) > 64 {
		err = errorf.E("subscription id must be length > 0 and <= 64")
		return
	}
	return &Result{subscription.MustNew(s), ev}, nil
}
func (en *Result) Label() S { return L }

func (en *Result) Write(w io.Writer) (err E) {
	var b B
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Result) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = en.Subscription.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = en.Event.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Result) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if en.Subscription, err = subscription.NewId(B{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.UnmarshalJSON(r); chk.E(err) {
		return
	}
	en.Event = event.New()
	if r, err = en.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseResult(b B) (t *Result, rem B, err E) {
	t = NewResult()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
