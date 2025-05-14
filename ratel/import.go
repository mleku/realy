package ratel

import (
	"bufio"
	"io"
	"os"
	"runtime/debug"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/log"
)

const maxLen = 500000000

// Import a collection of events in line structured minified JSON format (JSONL).
func (r *T) Import(rr io.Reader) (done chan struct{}) {
	done = make(chan struct{})
	r.Flatten = true
	// store to disk so we can return fast
	tmpPath := os.TempDir() + string(os.PathSeparator) + "realy"
	os.MkdirAll(tmpPath, 0700)
	tmp, err := os.CreateTemp(tmpPath, "")
	if chk.E(err) {
		return
	}
	log.I.F("buffering upload to %s", tmp.Name())
	if _, err = io.Copy(tmp, rr); chk.E(err) {
		return
	}
	if _, err = tmp.Seek(0, 0); chk.E(err) {
		return
	}
	log.I.F("finished buffering")
	go func() {
		defer tmp.Close()
		defer os.RemoveAll(tmpPath)
		defer close(done)
		scan := bufio.NewScanner(tmp)
		buf := make([]byte, maxLen)
		scan.Buffer(buf, maxLen)
		var count, total int
		for scan.Scan() {
			select {
			case <-r.Ctx.Done():
				log.I.F("context closed")
				return
			default:
			}
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
			if count%100 == 0 {
				log.I.F("received %d events", count)
				debug.FreeOSMemory()
			}
		}
		log.I.F("read %d bytes and saved %d events", total, count)
		err = scan.Err()
		if chk.E(err) {
		}
	}()
	// go func() {
	// 	if err = r.FulltextIndex(); chk.E(err) {
	// 		return
	// 	}
	// 	if err = r.LangIndex(); chk.E(err) {
	// 		return
	// 	}
	// }()
	return
}
