package ratel

import (
	"bytes"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
	"realy.lol/ratel/keys/fullpubkey"
	"realy.lol/ratel/keys/id"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/tag"
)

// GetIndexKeysForEvent generates all the index keys required to filter for
// events. evtSerial should be the output of Serial() which gets a unique,
// monotonic counter value for each new event.
func GetIndexKeysForEvent(ev *event.T, ser *serial.T) (keyz [][]byte) {

	var err error
	keyz = make([][]byte, 0, 18)
	ID := id.New(eventid.NewWith(ev.Id))
	CA := createdat.New(ev.CreatedAt)
	K := kinder.New(ev.Kind.ToU16())
	PK, _ := pubkey.New(ev.Pubkey)
	FID := fullid.New(eventid.NewWith(ev.Id))
	FPK := fullpubkey.New(ev.Pubkey)
	// indexes
	{ // ~ by id
		k := prefixes.Id.Key(ID, ser)
		// log.T.ToSliceOfBytes("id key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+date
		k := prefixes.Pubkey.Key(PK, CA, ser)
		// log.T.ToSliceOfBytes("pubkey + date key: %x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:17], k[17:])
		keyz = append(keyz, k)
	}
	{ // ~ by kind+date
		k := prefixes.Kind.Key(K, CA, ser)
		// log.T.ToSliceOfBytes("kind + date key: %x %0x %0x %0x",
		// 	k[0], k[1:3], k[3:11], k[11:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+kind+date
		k := prefixes.PubkeyKind.Key(PK, K, CA, ser)
		// log.T.ToSliceOfBytes("pubkey + kind + date key: %x %0x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:11], k[11:19], k[19:])
		keyz = append(keyz, k)
	}
	// ~ by tag value + date
	for i, t := range ev.Tags.ToSliceOfTags() {
		tsb := t.ToSliceOfBytes()
		if t.Len() < 2 {
			continue
		}
		tk, tv := tsb[0], tsb[1]
		if t.Len() < 2 ||
			// the tag is not a-zA-Z probably (this would permit arbitrary other single byte
			// chars)
			len(tk) != 1 ||
			// the second field is zero length
			len(tv) == 0 ||
			// the second field is more than 100 characters long
			len(tv) > 100 {
			// any of the above is true then the tag is not indexable
			continue
		}
		var firstIndex int
		var tt *tag.T
		for firstIndex, tt = range ev.Tags.ToSliceOfTags() {
			if tt.Len() >= 2 && bytes.Equal(tt.B(1), t.B(1)) {
				break
			}
		}
		if firstIndex != i {
			// duplicate
			continue
		}
		// create tags for e (event references) but we don't care about the optional third value
		// as it can't be searched for anyway (it's for clients to render threads)
		if bytes.Equal(tk, []byte("e")) {
			if len(tv) != 64 {
				continue
			}
			var ei []byte
			if ei, err = hex.DecAppend(ei, tv); chk.E(err) {
				continue
			}
			keyz = append(keyz, prefixes.TagEventId.Key(id.New(eventid.NewWith(ei)), ser))
			continue
		}
		// get key prefix (with full length) and offset where to write the last parts.
		prf, elems := index.P(0), []keys.Element(nil)
		if prf, elems, err = Create_a_Tag(string(tsb[0]),
			string(tv), CA, ser); chk.E(err) {
			// log.I.F("%v", t.ToStringSlice())
			return
		}
		k := prf.Key(elems...)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := prefixes.CreatedAt.Key(CA, ser)
		// log.T.ToSliceOfBytes("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	// { // Counter index - for storing last access time of events.
	// 	k := GetCounterKey(ser)
	// 	keyz = append(keyz, k)
	// }
	{ // - full Id index - enabling retrieving the event Id without unmarshalling the data
		k := prefixes.FullIndex.Key(ser, FID, FPK, CA)
		// log.T.ToSliceOfBytes("full id: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}
