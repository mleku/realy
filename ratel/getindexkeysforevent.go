package ratel

import (
	"bytes"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/event"
	"realy.mleku.dev/eventid"
	"realy.mleku.dev/log"
	"realy.mleku.dev/ratel/keys"
	"realy.mleku.dev/ratel/keys/createdat"
	"realy.mleku.dev/ratel/keys/fullid"
	"realy.mleku.dev/ratel/keys/fullpubkey"
	"realy.mleku.dev/ratel/keys/id"
	"realy.mleku.dev/ratel/keys/index"
	"realy.mleku.dev/ratel/keys/kinder"
	"realy.mleku.dev/ratel/keys/pubkey"
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/tag"
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
		// there is no value field
		if t.Len() < 2 ||
			// the tag is not a-zA-Z probably (this would permit arbitrary other
			// single byte chars)
			len(t.ToSliceOfBytes()[0]) != 1 ||
			// the second field is zero length
			len(t.ToSliceOfBytes()[1]) == 0 ||
			// the second field is more than 100 characters long
			len(t.ToSliceOfBytes()[1]) > 100 {
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
		// get key prefix (with full length) and offset where to write the last
		// parts
		prf, elems := index.P(0), []keys.Element(nil)
		if prf, elems, err = Create_a_Tag(string(t.ToSliceOfBytes()[0]),
			string(t.ToSliceOfBytes()[1]), CA,
			ser); chk.E(err) {
			log.I.F("%v", t.ToStringSlice())
			return
		}
		k := prf.Key(elems...)
		// log.T.ToSliceOfBytes("tag '%s': %s key %0x", t.ToSliceOfBytes()[0], t.ToSliceOfBytes()[1:], k)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := prefixes.CreatedAt.Key(CA, ser)
		// log.T.ToSliceOfBytes("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // Counter index - for storing last access time of events.
		k := GetCounterKey(ser)
		keyz = append(keyz, k)
	}
	{ // - full Id index - enabling retrieving the event Id without unmarshalling the data
		k := prefixes.FullIndex.Key(ser, FID, FPK, CA)
		// log.T.ToSliceOfBytes("full id: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}
