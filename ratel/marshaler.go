package ratel

import (
	"bytes"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/log"
)

func (r *T) Marshal(ev *event.T, dst []byte) (b []byte) {
	b = dst
	if r.Binary {
		buf := bytes.NewBuffer(dst)
		ev.MarshalBinary(buf)
		b = buf.Bytes()
	} else {
		b = ev.Marshal(dst)
	}
	return
}

func (r *T) Unmarshal(ev *event.T, b []byte) (rem []byte, err error) {
	if r.Binary {
		buf := bytes.NewBuffer(b)
		if err = ev.UnmarshalBinary(buf); chk.E(err) {
			return
		}
		rem = buf.Bytes()
	} else {
		if rem, err = ev.Unmarshal(b); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.T.S(rem)
		}
	}
	return
}
