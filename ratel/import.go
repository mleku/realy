package ratel

import (
	"bufio"
	"io"
	"time"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/event"
	"realy.mleku.dev/log"
)

const maxLen = 500000000

// Import a collection of events in line structured minified JSON format (JSONL).
func (r *T) Import(rr io.Reader) {
	r.Flatten = true
	var err error
	scan := bufio.NewScanner(rr)
	buf := make([]byte, maxLen)
	scan.Buffer(buf, maxLen)
	var count, total int
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
		b = nil
		ev = nil
		count++
		if count%1000 == 0 {
			log.I.F("received %d events", count)
			time.Sleep(time.Second)
		}
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
