package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"realy.lol/event"
	"realy.lol/interrupt"
	"realy.lol/units"
)

func main() {
	if len(os.Args) < 2 {
		log.F.Ln("input file argument required")
		os.Exit(1)
	}
	var fh *os.File
	var err er
	if fh, err = os.Open(os.Args[1]); chk.E(err) {
		os.Exit(1)
	}
	defer fh.Close()
	var read, unmar, ids, tobin, frombin, reser *os.File
	if read, err = os.OpenFile("read.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer read.Close()
	if unmar, err = os.OpenFile("fail_unmar.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer unmar.Close()
	if ids, err = os.OpenFile("fail_ids.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer ids.Close()
	if tobin, err = os.OpenFile("fail_tobin.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer tobin.Close()
	if frombin, err = os.OpenFile("fail_frombin.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer frombin.Close()
	if reser, err = os.OpenFile("fail_reser.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer reser.Close()
	interrupt.AddHandler(func() {
		unmar.Sync()
		ids.Sync()
		tobin.Sync()
		frombin.Sync()
		reser.Sync()
		os.Exit(0)
	})
	var progress, total no
	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make(by, units.Megabyte*8), units.Megabyte*8)
	bin := make(by, 0, units.Mb)
	cp := make(by, units.Mb)
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
		var rem by
		if rem, err = ev.Unmarshal(line); err != nil {
			// these two error types are fatal... json cannot have linebreak characters in
			// strings nor can events have keys that are other than the set defined in NIP-01.
			if err.Error() != "invalid character '\\n' in quoted string" &&
				!strings.HasPrefix(err.Error(), "invalid key,") {
				fmt.Fprintf(unmar, "%s\n", ev.Serialize())
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
		eq := equals(ev.ID, eh)
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
