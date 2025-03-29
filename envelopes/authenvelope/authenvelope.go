// Package authenvelope defines the auth challenge (relay message) and response
// (client message) of the NIP-42 authentication protocol.
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
	Challenge []byte
}

func NewChallenge() *Challenge { return &Challenge{} }
func NewChallengeWith[V string | []byte](challenge V) *Challenge {
	return &Challenge{[]byte(challenge)}
}
func (en *Challenge) Label() string { return L }

func (en *Challenge) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	log.T.F("writing out challenge envelope: '%s'", b)
	_, err = w.Write(b)
	return
}

func (en *Challenge) Marshal(dst []byte) (b []byte) {
	b = dst
	var err error
	b = envs.Marshal(b, L,
		func(bst []byte) (o []byte) {
			o = bst
			o = append(o, '"')
			o = text.NostrEscape(o, en.Challenge)
			o = append(o, '"')
			return
		})
	_ = err
	return
}

func (en *Challenge) Unmarshal(b []byte) (r []byte, err error) {
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

func ParseChallenge(b []byte) (t *Challenge, rem []byte, err error) {
	t = NewChallenge()
	if rem, err = t.Unmarshal(b); chk.E(err) {
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

func (en *Response) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

func (en *Response) Marshal(dst []byte) (b []byte) {
	var err error
	if en == nil {
		err = errorf.E("nil response")
		return
	}
	if en.Event == nil {
		err = errorf.E("nil event in response")
		return
	}
	b = dst
	b = envs.Marshal(b, L, en.Event.Marshal)
	_ = err
	return
}

func (en *Response) Unmarshal(b []byte) (r []byte, err error) {
	r = b
	// literally just unmarshal the event
	en.Event = event.New()
	if r, err = en.Event.Unmarshal(r); chk.E(err) {
		return
	}
	if r, err = envs.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

func ParseResponse(b []byte) (t *Response, rem []byte, err error) {
	t = NewResponse()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
