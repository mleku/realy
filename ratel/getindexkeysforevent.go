package ratel

import (
	"mleku.dev/event"
	"mleku.dev/eventid"
	"mleku.dev/ratel/keys"
	"mleku.dev/ratel/keys/createdat"
	"mleku.dev/ratel/keys/id"
	"mleku.dev/ratel/keys/index"
	"mleku.dev/ratel/keys/kinder"
	"mleku.dev/ratel/keys/pubkey"
	"mleku.dev/ratel/keys/serial"
	"mleku.dev/tag"
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
	// indexes
	{ // ~ by id
		k := index.Id.Key(ID, ser)
		// log.T.F("id key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+date
		k := index.Pubkey.Key(PK, CA, ser)
		// log.T.F("pubkey + date key: %x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:17], k[17:])
		keyz = append(keyz, k)
	}
	{ // ~ by kind+date
		k := index.Kind.Key(K, CA, ser)
		log.T.F("kind + date key: %x %0x %0x %0x",
			k[0], k[1:3], k[3:11], k[11:])
		keyz = append(keyz, k)
	}
	{ // ~ by pubkey+kind+date
		k := index.PubkeyKind.Key(PK, K, CA, ser)
		// log.T.F("pubkey + kind + date key: %x %0x %0x %0x %0x",
		// 	k[0], k[1:9], k[9:11], k[11:19], k[19:])
		keyz = append(keyz, k)
	}
	// ~ by tag value + date
	for i, t := range ev.Tags.T {
		// there is no value field
		if len(t.Field) < 2 ||
			// the tag is not a-zA-Z probably (this would permit arbitrary other
			// single byte chars)
			len(t.Field[0]) != 1 ||
			// the second field is zero length
			len(t.Field[1]) == 0 ||
			// the second field is more than 100 characters long
			len(t.Field[1]) > 100 {
			// any of the above is true then the tag is not indexable
			continue
		}
		var firstIndex int
		var tt *tag.T
		for firstIndex, tt = range ev.Tags.T {
			if len(tt.Field) >= 2 && equals(tt.Field[1], t.Field[1]) {
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
		if prf, elems, err = GetTagKeyElements(S(t.Field[1]), CA, ser); chk.E(err) {
			return
		}
		k := prf.Key(elems...)
		log.T.F("tag '%s': %s key %0x", t.Field[0], t.Field[1:], k)
		keyz = append(keyz, k)
	}
	{ // ~ by date only
		k := index.CreatedAt.Key(CA, ser)
		// log.T.F("date key: %x %0x %0x", k[0], k[1:9], k[9:])
		keyz = append(keyz, k)
	}
	return
}
