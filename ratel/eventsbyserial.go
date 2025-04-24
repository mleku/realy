package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.mleku.dev/eventidserial"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/ratel/keys"
	"realy.mleku.dev/ratel/keys/createdat"
	"realy.mleku.dev/ratel/keys/fullid"
	"realy.mleku.dev/ratel/keys/fullpubkey"
	"realy.mleku.dev/ratel/keys/index"
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/timestamp"
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
