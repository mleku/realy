package ratel

import (
	"realy.lol/event"
	"encoding/binary"
	"realy.lol/tags"
	"realy.lol/tag"
	"realy.lol/hex"
	"realy.lol/ec/schnorr"
	"strconv"
)

func (r *T) Marshal(ev *event.T, dst by) (b by) {
	if r.UseCompact {
		evp := *ev
		var err er
		if err = r.SubPubkeys(ev); chk.E(err) {
			return
		}
		b = evp.MarshalCompact(dst)
	} else {
		b = ev.Marshal(dst)
	}
	return
}

// SubPubkeys substitutes the pubkeys in an event for the index keys, used as
// part of the marshal process.
func (r *T) SubPubkeys(ev *event.T) (err er) {
	// first substitute the event pubkey, this is encoded in raw bytes
	var ser u64
	if ser, err = r.GetPubkeyIndex(ev.PubKey); chk.E(err) {
		return
	}
	b := make(by, 8)
	binary.BigEndian.PutUint64(b, ser)
	// trim off the leading zeroes, they waste space in the compact json encoding.
	for i := range b {
		if b[i] != 0 {
			b = b[i:]
			break
		}
	}
	ev.PubKey = b
	// next search the p tags
	tgs := ev.Tags.F()
	for i := range tgs {
		if equals(tgs[i].Key(), by("p")) {
			bs := tgs[i].BS()
			// the value is hex encoded
			pkb := make(by, schnorr.PubKeyBytesLen)
			if _, err = hex.DecBytes(pkb, bs[tag.Value]); chk.E(err) {
				return
			}
			// get the index serial for the pubkey (this creates the index if not exists)
			if ser, err = r.GetPubkeyIndex(pkb); chk.E(err) {
				return
			}
			// convert the serial into hex
			idb := by(strconv.FormatUint(ser, 16))
			// zero pad to ensure the hex decoder can convert to the raw bytes
			if len(idb)%2 == 1 {
				idb = append(by("0"), idb...)
			}
			bs[tag.Value] = idb
			tgs[i] = tag.New(bs...)
		}
	}
	ev.Tags = tags.New(tgs...)
	return
}

func (r *T) Unmarshal(ev *event.T, evb by) (rem by, err er) {
	if r.UseCompact {
		if rem, err = ev.UnmarshalCompact(evb); chk.E(err) {
			ev = nil
			evb = evb[:0]
			return
		}
		if err = r.SubIndexes(ev); chk.E(err) {
			return
		}
	} else {
		if rem, err = ev.Unmarshal(evb); chk.E(err) {
			ev = nil
			evb = evb[:0]
			return
		}
	}
	return
}

// SubIndexes substitutes the pubkeys for the indexes in an event, part of the
// unmarshal process.
func (r *T) SubIndexes(ev *event.T) (err er) {
	// first sub the event pubkey
	var idpk by
	if idpk, err = r.GetIndexPubkey(ev.PubKey); chk.E(err) {
		return
	}
	ev.PubKey = idpk
	// next the tags
	tgs := ev.Tags.F()
	for i := range tgs {
		if equals(tgs[i].Key(), by("p")) {
			bs := tgs[i].BS()
			var ptpk by
			b := bs[tag.Value]
			if ptpk, err = r.GetIndexPubkey(b); chk.E(err) {
				return
			}
			b = b[:0]
			bs[tag.Value] = hex.EncAppend(b, ptpk)
		}
	}
	return
}
