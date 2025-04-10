package ratel

import (
	"realy.mleku.dev/event"
)

// Unmarshal an event from bytes, using compact encoding if configured.
func (r *T) Unmarshal(ev *event.T, evb []byte) (rem []byte, err error) {
	if r.UseCompact {
		if rem, err = ev.UnmarshalCompact(evb); chk.E(err) {
			ev = nil
			evb = evb[:0]
			return
		}
	} else {
		if rem, err = ev.Unmarshal(evb); chk.E(err) {
			ev = nil
			evb = evb[:0]
			return
		}
	}
	return
}

// Marshal an event using compact encoding if configured.
func (r *T) Marshal(ev *event.T, dst []byte) (b []byte) {
	if r.UseCompact {
		b = ev.MarshalCompact(dst)
	} else {
		b = ev.Marshal(dst)
	}
	return
}
