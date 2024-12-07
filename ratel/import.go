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
	buf := make(by, maxLen)
	scan.Buffer(buf, maxLen)
	var err er
	var count no
	for scan.Scan() {
		b := scan.Bytes()
		ev := &event.T{}
		if _, err = ev.Unmarshal(b); err != nil {
			continue
		}
		if err = r.SaveEvent(r.Ctx, ev); err != nil {
			continue
		}
		count++
		if count > 0 && count%1000 == 0 {
			chk.T(r.DB.Sync())
			chk.T(r.DB.RunValueLogGC(0.5))
		}
	}
	err = scan.Err()
	if chk.E(err) {
	}
	return
}
