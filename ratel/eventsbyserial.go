package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/eventidserial"
	"realy.lol/hex"
	"realy.lol/ratel/keys"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
	"realy.lol/ratel/keys/fullpubkey"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/timestamp"
)

func (r *T) EventIdsBySerial(start uint64, count int) (evs []eventidserial.E,
	err error) {
	err = r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: prefixes.FullIndex.Key()})
		defer it.Close()
		ser := serial.New(serial.Make(start))
		prf := prefixes.FullIndex.Key(ser)
		it.Seek(prf)
		if count > 1000 {
			count = 1000
		}
		var counter int
		for ; counter < count && it.Valid(); it.Next() {
			counter++
			item := it.Item()
			k := item.KeyCopy(nil)
			id := fullid.New()
			ts := createdat.New(timestamp.New())
			pk := fullpubkey.New()
			keys.Read(k, index.New(0), ser, id, pk, ts)
			// counter++
			evs = append(evs, eventidserial.E{
				Serial:  ser.Uint64(),
				EventId: hex.Enc(id.Val),
			})
		}
		return
	})
	return
}
