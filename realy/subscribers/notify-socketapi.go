package subscribers

import (
	"bytes"

	"realy.mleku.dev/event"
	"realy.mleku.dev/tag"
)

func (s *S) NotifyHTTPAPI(authRequired, publicReadable bool, ev *event.T) {
	s.HMx.Lock()
	var subs []*H
	for sub := range s.HMap {
		// check if the subscription's subscriber is still alive
		select {
		case <-sub.Ctx.Done():
			subs = append(subs, sub)
		default:
		}
	}
	for _, sub := range subs {
		delete(s.HMap, sub)
	}
	subs = subs[:0]
	for sub := range s.HMap {
		// if auth required, check the subscription pubkey matches
		if !publicReadable {
			if authRequired && len(sub.Pubkey) == 0 {
				continue
			}
		}
		// if the filter doesn't match, skip
		if !sub.Filter.Matches(ev) {
			continue
		}
		// if the filter is privileged and the user doesn't have matching auth, skip
		if ev.Kind.IsPrivileged() {
			ab := sub.Pubkey
			var containsPubkey bool
			if ev.Tags != nil {
				containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(ab))
			}
			if !bytes.Equal(ev.Pubkey, ab) || containsPubkey {
				continue
			}
		}
		// send the event to the subscriber
		sub.Receiver <- ev
	}
	s.HMx.Unlock()
}
