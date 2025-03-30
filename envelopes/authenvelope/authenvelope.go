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

// Challenge is the relay-sent message containing a relay-chosen random string
// to prevent replay attacks on NIP-42 authentication.
type Challenge struct {
	Challenge []byte
}

var _ codec.Envelope = (*Challenge)(nil)

// NewChallenge creates a new empty Challenge.
func NewChallenge() *Challenge { return &Challenge{} }

// NewChallengeWith creates a new Challenge with provided bytes.
func NewChallengeWith[V string | []byte](challenge V) *Challenge {
	return &Challenge{[]byte(challenge)}
}

// Label returns the label of a Challenge envelope.
func (en *Challenge) Label() string { return L }

// Write the Challenge to a provided io.Writer.
func (en *Challenge) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	log.T.F("writing out challenge envelope: '%s'", b)
	_, err = w.Write(b)
	return
}

// Marshal a Challenge to minified JSON. Note that this ensures correct string
// escaping on the challenge field.
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

// Unmarshal a Challenge from minified JSON. Note that this ensures the
// challenge string was correctly escaped by NIP-01 escaping rules.
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

// ParseChallenge reads a Challenge encoded in minified JSON and unpacks it to
// the runtime format.
func ParseChallenge(b []byte) (t *Challenge, rem []byte, err error) {
	t = NewChallenge()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}

// Response is a client-side envelope containing the signed event bearing the
// relay's URL and Challenge string.
type Response struct {
	Event *event.T
}

var _ codec.Envelope = (*Response)(nil)

// NewResponse creates a new empty Response.
func NewResponse() *Response { return &Response{} }

// NewResponseWith creates a new Response with a provided event.T.
func NewResponseWith(event *event.T) *Response { return &Response{Event: event} }

// Label returns the label of a auth Response envelope.
func (en *Response) Label() string { return L }

// Write the Response to a provided io.Writer.
func (en *Response) Write(w io.Writer) (err error) {
	var b []byte
	b = en.Marshal(b)
	_, err = w.Write(b)
	return
}

// Marshal a Response to minified JSON. Note that this ensures correct string
// escaping on the challenge field.
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

// Unmarshal a Response from minified JSON. Note that this ensures the
// challenge string was correctly escaped by NIP-01 escaping rules.
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

// ParseResponse reads a Response encoded in minified JSON and unpacks it to
// the runtime format.
func ParseResponse(b []byte) (t *Response, rem []byte, err error) {
	t = NewResponse()
	if rem, err = t.Unmarshal(b); chk.E(err) {
		return
	}
	return
}
