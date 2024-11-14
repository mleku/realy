package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"realy.lol/event"
	"realy.lol/interrupt"
	"realy.lol/lol"
	"realy.lol/qu"
	"realy.lol/units"
)

func main() {
	lol.SetLogLevel("trace")
	if len(os.Args) < 2 {
		log.F.Ln("filename of .jsonl file for testing the codecs is required")
		os.Exit(1)
	}
	fh, err := os.Open(os.Args[1])
	if err != nil {
		log.E.F("error opening file: %v", err)
		os.Exit(1)
	}
	defer fh.Close()
	var skip N
	if len(os.Args) > 2 {
		if skip, err = strconv.Atoi(os.Args[2]); err != nil {
			log.E.F("error parsing flag: %v", err)
		}
		skip *= units.Gb
	}
	if _, err = fh.Seek(int64(skip), 0); chk.E(err) {
		os.Exit(1)
	}
	// var taglog *os.File
	// if taglog, err = os.OpenFile("./failedtags.jsonl", os.O_CREATE|os.O_TRUNC,
	// 	0666); chk.E(err) {
	// 	os.Exit(1)
	// }
	// defer taglog.Close()
	b := make(B, 1)
	bin := make(B, 0, units.Mb)
	var count, total, pos int
	total = skip
	line := make([]byte, 0, units.Mb)
	quit := qu.T()
	interrupt.AddHandler(func() { quit.Q() })
out:
	for {
		select {
		case <-quit:
			break out
		default:
		}
		_, err = fh.Read(b)
		if err != nil {
			if err != io.EOF {
				break
			}
		}
		pos++
		if b[0] == '\n' {
			if total%1000 == 0 {
				log.I.F("line %d megabyte %d gigabyte %d", total, pos/units.Megabyte,
					pos/units.Gigabyte)
			}
			total++
			cp := make(B, len(line))
			copy(cp, line)
			ev := event.T{}
			var rem B
			if rem, err = ev.UnmarshalJSON(line); err != nil {
				if err.Error() != "invalid character '\\n' in quoted string" &&
					!strings.HasPrefix(err.Error(), "invalid key,") {
					fmt.Fprintf(os.Stdout, "%s\n", ev.Serialize())
					// log.E.F("error unmarshaling line: '%s'\n%s", err.Error(), cp)
				}
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
				// can't use this event because its ID is incorrectly computed
				if !equals(cp, ev.Serialize()) {
					fmt.Fprintf(os.Stdout, "%s\n", ev.Serialize())
					// log.W.F("failed encode to get correct ID\n%s\n%s", cp, ev.Serialize())
				}
				line = line[:0]
				continue
			}
			if bin, err = ev.MarshalBinary(bin); err != nil {
				if strings.HasPrefix(err.Error(), "encoding/hex: invalid byte:") {
					// has a hashtag in a p tag or garbage in an e tag
					line = line[:0]
					bin = bin[:0]
					continue
				}
				fmt.Fprintf(os.Stdout, "%s\n", ev.Serialize())
				// log.E.F("failed encode to binary\n%s\n%s", err.Error(), cp)
				line = line[:0]
				bin = bin[:0]
				continue
			}
			ev2 := &event.T{}
			if rem, err = ev2.UnmarshalBinary(bin); err != nil {
				// log.I.F("%s", cp)
				// log.I.S(bin, rem)
				// fmt.Fprintf(os.Stderr, "failed back from binary\n%s", cp)
				// panic(err.Error())
				// log the event that triggers the binary tag codec error
				fmt.Fprintf(os.Stdout, "%s\n", ev.Serialize())
				bin = bin[:0]
				line = line[:0]
				continue
			}
			if !equals(ev.Serialize(), ev2.Serialize()) {
				// log.E.F("%s", cp)
				// log.I.S(ev, bin, ev2, rem, len(ev.Content))
				// fmt.Fprintf(os.Stderr,
				// 	"decoded back from binary not same as original decoded\n\n"+
				// 		"original:\n\n%s\n\n"+
				// 		"decoded from JSON:\n\n%s\n\n"+
				// 		"encoded to binary and decoded:\n\n%s\n\n",
				// 	cp, ev.Serialize(), ev2.Serialize())
				// panic(err.Error())
				// log the event that triggers the binary tag codec error
				fmt.Fprintf(os.Stdout, "%s\n", ev.Serialize())
				bin = bin[:0]
				line = line[:0]
				continue
			}
			count++
			// if count%100 == 0 {
			// 	log.I.Ln("valid", count, "total", total)
			// }
			bin = bin[:0]
			line = line[:0]
		} else {
			line = append(line, b[0])
		}
	}
	log.I.Ln("valid", count, "total", total)
}
