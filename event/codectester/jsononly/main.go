package main

import (
	"bufio"
	"fmt"
	"os"
	"realy.lol/event"
	"realy.lol/interrupt"
	"realy.lol/units"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.F.Ln("input file argument required")
		os.Exit(1)
	}
	var fh *os.File
	var err E
	if fh, err = os.Open(os.Args[1]); chk.E(err) {
		os.Exit(1)
	}
	defer fh.Close()
	var (
		read    *os.File
		unmar   *os.File
		ids     *os.File
		back    *os.File
		reser   *os.File
		tobin   *os.File
		frombin *os.File
		rebin   *os.File
	)
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
	if back, err = os.OpenFile("fail_back.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer back.Close()
	if reser, err = os.OpenFile("fail_reser.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer reser.Close()
	if reser, err = os.OpenFile("fail_reser.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer reser.Close()
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
	if rebin, err = os.OpenFile("fail_rebin.jsonl", os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600); chk.E(err) {
		return
	}
	defer rebin.Close()
	interrupt.AddHandler(func() {
		read.Sync()
		unmar.Sync()
		ids.Sync()
		back.Sync()
		reser.Sync()
		tobin.Sync()
		frombin.Sync()
		rebin.Sync()
		os.Exit(0)
	})
	var progress, total N
	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make(B, units.Megabyte*8), units.Megabyte*8)
	bin := make(B, 0, units.Mb*8)
	cp := make(B, units.Mb*8)
	start := time.Now()
	for scanner.Scan() {
		progress++
		if progress%1000 == 0 {
			tot := float64(total) / float64(units.Megabyte)
			log.I.F("progress: line %d megabytes %f %f mb/s", progress,
				tot, tot/float64(time.Now().Sub(start).Seconds()))
		}
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
		var rem B
		if rem, err = ev.UnmarshalJSON(line); err != nil {
			// these two error types are fatal... json cannot have linebreak characters in
			// strings nor can events have keys that are other than the set defined in NIP-01.
			//if err.Error() != "invalid character '\\n' in quoted string" &&
			//	!strings.HasPrefix(err.Error(), "invalid key,") {
			fmt.Fprintf(unmar, "%s\n%s\n", cp, ev.Serialize())
			//log.E.F("error unmarshaling line: '%s'\n%s", err.Error(), cp)
			//}
			line = line[:0]
			continue
		}
		if len(rem) > 0 {
			log.I.F("remainder:\n%s", rem)
		}
		can := ev.ToCanonical()
		eh := event.Hash(can)
		eq := equals(ev.ID, eh)
		if !eq {
			ev.ID = eh
			_, err = fmt.Fprintf(ids, "%s\n%s\n", cp, ev.Serialize())
			line = line[:0]
			continue
		}
		ev2 := &event.T{}
		evSer := ev.Serialize()
		//log.I.F("%s", evSer)
		if rem, err = ev2.UnmarshalJSON(evSer); err != nil {
			// these two error types are fatal... json cannot have linebreak characters in
			// strings nor can events have keys that are other than the set defined in NIP-01.
			//if err.Error() != "invalid character '\\n' in quoted string" &&
			//	!strings.HasPrefix(err.Error(), "invalid key,") {
			fmt.Fprintf(reser, "%s\n%s\n", cp, ev2.Serialize())
			//log.E.F("error unmarshaling line: '%s'\n%s", err.Error(), cp)
			//}
			line = line[:0]
			continue
		}
		if len(rem) > 0 {
			log.I.F("remainder:\n%s", rem)
		}
		if !equals(ev.Serialize(), ev2.Serialize()) {
			_, err = fmt.Fprintf(reser, "%s\n%s\n%s\n\n", cp, ev.Serialize(), ev2.Serialize())
			line = line[:0]
			continue
		}
		if bin, err = ev2.MarshalBinary(bin); err != nil {
			_, err = fmt.Fprintf(tobin, "%s\n%s\n\n", cp, ev2.Serialize())
			line = line[:0]
			continue
		}
		ev3 := &event.T{}
		if rem, err = ev3.UnmarshalBinary(bin); err != nil {
			_, err = fmt.Fprintf(frombin, "%s\n%s\n\n", cp, ev2.Serialize())
			line = line[:0]
			continue
		}
		if !equals(ev2.Serialize(), ev3.Serialize()) {
			//log.I.S(ev2, ev3)
			_, err = fmt.Fprintf(rebin, "%s\n%s\n%s\n\n", cp, ev2.Serialize(), ev3.Serialize())
			line = line[:0]
			continue
		}
	}
	chk.E(scanner.Err())
}
