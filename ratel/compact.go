package ratel

import (
	"realy.lol/event"
)

func (r *T) Unmarshal(ev *event.T, evb by) (rem by, err er) {
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

func (r *T) Marshal(ev *event.T, dst by) (b by) {
	if r.UseCompact {
		b = ev.MarshalCompact(dst)
	} else {
		b = ev.Marshal(dst)
	}
	return
}
