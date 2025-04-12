package subscribers

import (
	"bytes"

	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/event"
	"realy.mleku.dev/tag"
)

func (s *S) NotifySocketAPI(authRequired, publicReadable bool, ev *event.T) {
	var err error
	s.WsMx.Lock()
	for ws, subs := range s.WsMap {
		for id, listener := range subs {
			if !publicReadable {
				if authRequired && !ws.IsAuthed() {
					continue
				}
			}
			if !listener.filters.Match(ev) {
				continue
			}
			if ev.Kind.IsPrivileged() {
				ab := ws.AuthedBytes()
				var containsPubkey bool
				if ev.Tags != nil {
					containsPubkey = ev.Tags.ContainsAny([]byte{'p'}, tag.New(ab))
				}
				if !bytes.Equal(ev.Pubkey, ab) || containsPubkey {
					if ab == nil {
						continue
					}
					log.I.F("authed user %0x not privileged to receive event\n%s",
						ab, ev.Serialize())
					continue
				}
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(id, ev); chk.E(err) {
				continue
			}
			if err = res.Write(ws); chk.E(err) {
				continue
			}
		}
	}
	s.WsMx.Unlock()

}
