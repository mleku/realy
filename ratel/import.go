package ratel

import (
	"bufio"
	"io"

	"realy.lol/event"
)

const maxLen = 500000000

// Import accepts an event
func (r *T) Import(rr io.Reader) {
	r.Flatten = true
	scan := bufio.NewScanner(rr)
	buf := make(B, maxLen)
	scan.Buffer(buf, maxLen)
	var err E
	var count N
	for scan.Scan() {
		b := scan.Bytes()
		// if len(b) > 8192 {
		// 	log.I.F("saving,%s", b)
		// }
		ev := &event.T{}
		if _, err = ev.UnmarshalJSON(b); err != nil {
			continue
		}
		if err = r.SaveEvent(r.Ctx, ev); err != nil {
			continue
		}
		count++
		if count > 0 && count%100 == 0 {
			chk.T(r.DB.Sync())
			log.I.F("imported 10000/%d events, running GC on new data", count)
			chk.T(r.DB.RunValueLogGC(0.5))
		}
	}
	err = scan.Err()
	if chk.E(err) {
	}
	return
}
