package event

import (
	"encoding/base64"
)

// MarshalWireCompact encodes an event as the canonical form wrapped in an array
// with the signature encoded in raw Base64 URL (86 bytes instead of 128 of hex).
func (ev *T) MarshalWireCompact(dst []byte) (b []byte) {
	b = dst
	b = append(b, '[')
	b = ev.ToCanonical(b)
	b = append(b, ',', '"')
	l := len(b)
	b = append(b, make([]byte, 86)...)
	base64.RawURLEncoding.Encode(b[l:], ev.Sig)
	b = append(b, '"', ']')
	return
}

// UnmarshalWireCompact decodes an event encoded in minified Wire Compact form -
// with an enclosing array around the canonical form of the event with the
// signature encoded in Base64 URL (86 bytes instead of 128 of hex).
func (ev *T) UnmarshalWireCompact(b []byte) (rem []byte, err error) {
	startEv := 1
	endSig := len(b) - 2
	startSig := endSig - 86
	endEv := startSig - 2
	evB := b[startEv:endEv]
	id := Hash(evB)
	if rem, err = ev.FromCanonical(evB); chk.E(err) {
		return
	}
	sigB := make([]byte, 64)
	_, err = base64.RawURLEncoding.Decode(sigB, b[startSig:endSig])
	ev.Sig = sigB
	ev.Id = id
	return
}
