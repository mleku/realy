package ratel

import (
	"bytes"

	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
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
	ID := id.New(eventid.NewWith(ev.ID))
	CA := createdat.New(ev.CreatedAt)
	K := kinder.New(ev.Kind.ToU16())
	PK, _ := pubkey.New(ev.PubKey)
	FID := fullid.New(eventid.NewWith(ev.ID))
	// indexes
	{ // ~ by id
		k := prefixes.Id.Key(ID, ser)
		// log.T.F("id key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+date
		k := prefixes.Pubkey.Key(PK, CA, ser)
		// log.T.F("pubkey + date key: %x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:17], k[17:])
		keyz = append(keyz, k)
	}
	{ // ~ by kind+date
		k := prefixes.Kind.Key(K, CA, ser)
		// log.T.F("kind + date key: %x %0x %0x %0x",
		// 	k[0], k[1:3], k[3:11], k[11:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+kind+date
		k := prefixes.PubkeyKind.Key(PK, K, CA, ser)
		// log.T.F("pubkey + kind + date key: %x %0x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:11], k[11:19], k[19:])
		keyz = append(keyz, k)
	}
	// ~ by tag value + date
	for i, t := range ev.Tags.Value() {
		// there is no value field
		if t.Len() < 2 ||
			// the tag is not a-zA-Z probably (this would permit arbitrary other
			// single byte chars)
			len(t.F()[0]) != 1 ||
			// the second field is zero length
			len(t.F()[1]) == 0 ||
			// the second field is more than 100 characters long
			len(t.F()[1]) > 100 {
			// any of the above is true then the tag is not indexable
			continue
		}
		var firstIndex int
		var tt *tag.T
		for firstIndex, tt = range ev.Tags.Value() {
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
		if prf, elems, err = GetTagKeyElements(string(t.F()[0]), string(t.F()[1]), CA, ser); chk.E(err) {
			log.I.F("%v", t.ToStringSlice())
			return
		}
		k := prf.Key(elems...)
		// log.T.F("tag '%s': %s key %0x", t.F()[0], t.F()[1:], k)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := prefixes.CreatedAt.Key(CA, ser)
		// log.T.F("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // Counter index - for storing last access time of events.
		k := GetCounterKey(ser)
		keyz = append(keyz, k)
	}
	{ // - full ID index - enabling retrieving the event ID without unmarshalling the data
		k := prefixes.FullIdIndex.Key(FID, CA, ser)
		// log.T.F("full id: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}
