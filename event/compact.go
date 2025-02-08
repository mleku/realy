package event

import (
	"realy.lol/ec/schnorr"
)

// MarshalCompact encodes an event as the canonical form followed by the raw binary
// signature (64 bytes) which hashes to form the ID, thus a compact form for the
// database that is smaller and fast to decode.
func (ev *T) MarshalCompact(dst []byte) (b []byte) {
	b = dst
	b = ev.ToCanonical(b)
	b = append(b, ev.Sig...)
	return
}

func (ev *T) UnmarshalCompact(b []byte) (rem []byte, err error) {
	rem = b
	end := len(rem) - schnorr.SignatureSize
	id := Hash(rem[:end])
	if rem, err = ev.FromCanonical(b); chk.E(err) {
		return
	}
	ev.Sig = rem[:schnorr.SignatureSize]
	ev.ID = id
	rem = rem[schnorr.SignatureSize:]
	return
}
