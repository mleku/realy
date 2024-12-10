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
	var err er
	scan := bufio.NewScanner(rr)
	buf := make(by, maxLen)
	scan.Buffer(buf, maxLen)
	var count, total no
	for scan.Scan() {
		b := scan.Bytes()
		total += len(b) + 1
		if len(b) < 1 {
			continue
		}
		ev := &event.T{}
		if _, err = ev.Unmarshal(b); err != nil {
			continue
		}
		if err = r.SaveEvent(r.Ctx, ev); err != nil {
			continue
		}
		count++
		if count > 0 && count%10000 == 0 {
			chk.T(r.DB.Sync())
			chk.T(r.DB.RunValueLogGC(0.5))
		}
	}
	log.I.F("read %d bytes and saved %d events", total, count)
	err = scan.Err()
	if chk.E(err) {
	}
	return
}
