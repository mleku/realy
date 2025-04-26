// Package main is a tester that reads in a provided JSON line structured
// (.jsonl) document containing a set of events and attempts to parse them and
// prints out the events that failed various steps in the encode/decode process.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/interrupt"
	"realy.lol/log"
	"realy.lol/units"
)

func main() {
	if len(os.Args) < 2 {
		log.F.Ln("input file argument required")
		os.Exit(1)
	}
	var fh *os.File
	var err error
	if fh, err = os.Open(os.Args[1]); chk.E(err) {
		os.Exit(1)
	}
	defer func() { _ = fh.Close() }()
	var read, unmar, ids, tobin, frombin, reser *os.File
	if read, err = os.OpenFile("read.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = read.Close() }()
	if unmar, err = os.OpenFile("fail_unmar.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = unmar.Close() }()
	if ids, err = os.OpenFile("fail_ids.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = ids.Close() }()
	if tobin, err = os.OpenFile("fail_tobin.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = tobin.Close() }()
	if frombin, err = os.OpenFile("fail_frombin.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = frombin.Close() }()
	if reser, err = os.OpenFile("fail_reser.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer func() { _ = reser.Close() }()
	interrupt.AddHandler(func() {
		_ = unmar.Sync()
		_ = ids.Sync()
		_ = tobin.Sync()
		_ = frombin.Sync()
		_ = reser.Sync()
		os.Exit(0)
	})
	var progress, total int
	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, units.Megabyte*8), units.Megabyte*8)
	bin := make([]byte, 0, units.Mb)
	cp := make([]byte, units.Mb)
	for scanner.Scan() {
		cp = cp[:0]
		bin = bin[:0]
		line := scanner.Bytes()
		_, err = fmt.Fprintf(read, "%s\n", line)
		if chk.E(err) {
			panic(err)
		}
		total += len(line) + 1
		cp = append(cp, line...)
		ev := event.T{}
		var rem []byte
		if rem, err = ev.Unmarshal(line); err != nil {
			// these two error types are fatal... json cannot have linebreak characters in
			// strings nor can events have keys that are other than the set defined in NIP-01.
			if err.Error() != "invalid character '\\n' in quoted string" &&
				!strings.HasPrefix(err.Error(), "invalid key,") {
				_, _ = fmt.Fprintf(unmar, "%s\n", ev.Serialize())
				log.E.F("error unmarshaling line: '%s'\n%s", err.Error(), cp)
			}
			line = line[:0]
			continue
		}
		if len(rem) > 0 {
			log.I.F("remainder:\n%s", rem)
		}
		can := ev.ToCanonical(nil)
		eh := event.Hash(can)
		eq := bytes.Equal(ev.Id, eh)
		if !eq {
			_, err = fmt.Fprintf(ids, "%s\n", ev.Serialize())
			if chk.E(err) {
				panic(err)
			}
			line = line[:0]
			continue
		}
		// if bin, err = ev.MarshalBinary(bin); err != nil {
		//	_, err = fmt.Fprintf(tobin, "%s\n", ev.Serialize())
		//	if chk.E(err) {
		//		panic(err)
		//	}
		//	line = line[:0]
		//	continue
		// }
		// ev2 := &event.T{}
		// if rem, err = ev2.UnmarshalBinary(bin); err != nil {
		//	_, err = fmt.Fprintf(frombin, "%s\n", ev.Serialize())
		//	if chk.E(err) {
		//		panic(err)
		//	}
		//	line = line[:0]
		//	continue
		// }
		// if !equals(ev.Serialize(), ev2.Serialize()) {
		//	_, err = fmt.Fprintf(reser, "%s\n", ev.Serialize())
		//	if chk.E(err) {
		//		panic(err)
		//	}
		//	line = line[:0]
		//	continue
		// }
		progress++
		if progress%1000 == 0 {
			log.I.F("progress: line %d megabytes %f", progress,
				float64(total)/float64(units.Megabyte))
		}
	}
	chk.E(scanner.Err())
}
