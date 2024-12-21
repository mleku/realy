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
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *Submission) Marshal(dst by) (b by) {
	var err er
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst by) (o by) {
			o = bst
			o = en.T.Marshal(o)
			return
		})
	_ = err
	return
}

func (en *Submission) Unmarshal(b by) (r by, err er) {
	r = b
	en.T = event.New()
	if r, err = en.T.Unmarshal(r); chk.T(err) {
		return
	}
	r = en.T.Marshal(nil)
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseSubmission(b by) (t *Submission, rem by, err er) {
	t = NewSubmission()
	if rem, err = t.Unmarshal(b); chk.E(err) {
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

func NewResult() *Result { return &Result{Event: &event.T{}} }

func NewResultWith[V st | by](s V, ev *event.T) (res *Result, err er) {
	if len(s) < 0 || len(s) > 64 {
		err = errorf.E("subscription id must be length > 0 and <= 64")
		return
	}
	return &Result{subscription.MustNew(s), ev}, nil
}

func (en *Result) Label() st { return L }

func (en *Result) Write(w io.Writer) (err er) {
	_, err = w.Write(en.Marshal(nil))
	return
}

func (en *Result) Marshal(dst by) (b by) {
	var err er
	b = dst
	b = envelopes.Marshal(b, L,
		func(bst by) (o by) {
			o = bst
			o = en.Subscription.Marshal(o)
			o = append(o, ',')
			o = en.Event.Marshal(o)
			return
		})
	_ = err
	return
}

func (en *Result) Unmarshal(b by) (r by, err er) {
	r = b
	if en.Subscription, err = subscription.NewId(by{0}); chk.E(err) {
		return
	}
	if r, err = en.Subscription.Unmarshal(r); chk.E(err) {
		return
	}
	en.Event = event.New()
	if r, err = en.Event.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	// log.I.S(en.Event)
	return
}

func ParseResult(b by) (t *Result, rem by, err er) {
	t = NewResult()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
