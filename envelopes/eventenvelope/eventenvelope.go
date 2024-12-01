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

func (en *Submission) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Submission) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst by) (o by, err er) {
			o = bst
			if o, err = en.T.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (en *Submission) UnmarshalJSON(b by) (r by, err er) {
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

func ParseSubmission(b by) (t *Submission, rem by, err er) {
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
func NewResultWith[V st | by](s V, ev *event.T) (res *Result, err er) {
	if len(s) < 0 || len(s) > 64 {
		err = errorf.E("subscription id must be length > 0 and <= 64")
		return
	}
	return &Result{subscription.MustNew(s), ev}, nil
}
func (en *Result) Label() st { return L }

func (en *Result) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Result) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst by) (o by, err er) {
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

func (en *Result) UnmarshalJSON(b by) (r by, err er) {
	r = b
	if en.Subscription, err = subscription.NewId(by{0}); chk.E(err) {
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

func ParseResult(b by) (t *Result, rem by, err er) {
	t = NewResult()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
