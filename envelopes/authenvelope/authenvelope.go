package authenvelope

import (
	"io"

	"realy.lol/codec"
	envs "realy.lol/envelopes"
	"realy.lol/event"
	"realy.lol/text"
)

const L = "AUTH"

type Challenge struct {
	Challenge by
}

func NewChallenge() *Challenge                           { return &Challenge{} }
func NewChallengeWith[V st | by](challenge V) *Challenge { return &Challenge{by(challenge)} }
func (en *Challenge) Label() string                      { return L }

func (en *Challenge) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	log.T.F("writing out challenge envelope: '%s'", b)
	_, err = w.Write(b)
	return
}

func (en *Challenge) MarshalJSON(dst by) (b by, err er) {
	b = dst
	b, err = envs.Marshal(b, L,
		func(bst by) (o by, err er) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Challenge)
			o = append(o, '"')
			return
		})
	return
}

func (en *Challenge) UnmarshalJSON(b by) (r by, err er) {
	r = b
	if en.Challenge, r, err = text.UnmarshalQuoted(r); chk.E(err) {
		return
	}
	for ; len(r) >= 0; r = r[1:] {
		if r[0] == ']' {
			r = r[:0]
			return
		}
	}
	return
}

func ParseChallenge(b by) (t *Challenge, rem by, err er) {
	t = NewChallenge()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}

type Response struct {
	Event *event.T
}

var _ codec.Envelope = (*Response)(nil)

func NewResponse() *Response                   { return &Response{} }
func NewResponseWith(event *event.T) *Response { return &Response{Event: event} }
func (en *Response) Label() string             { return L }

func (en *Response) Write(w io.Writer) (err er) {
	var b by
	if b, err = en.MarshalJSON(b); chk.E(err) {
		return
	}
	_, err = w.Write(b)
	return
}

func (en *Response) MarshalJSON(dst by) (b by, err er) {
	if en == nil {
		err = errorf.E("nil response")
		return
	}
	if en.Event == nil {
		err = errorf.E("nil event in response")
		return
	}
	b = dst
	b, err = envs.Marshal(b, L, en.Event.MarshalJSON)
	return
}

func (en *Response) UnmarshalJSON(b by) (r by, err er) {
	r = b
	// literally just unmarshal the event
	en.Event = event.New()
	if r, err = en.Event.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envs.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseResponse(b by) (t *Response, rem by, err er) {
	t = NewResponse()
	if rem, err = t.UnmarshalJSON(b); chk.E(err) {
		return
	}
	return
}
