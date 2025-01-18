package ratel

import (
	"github.com/dgraph-io/badger/v4"
	"realy.lol/ec/schnorr"
	"realy.lol/ratel/prefixes"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"encoding/binary"
)

// GetPubkeyIndex returns the index key for a pubkey, returning the one that
// exists if found, or generating a new one and returns also the serial bytes
// and serial uint64 in both cases.
func (r *T) GetPubkeyIndex(pk by) (ser uint64, err er) {
	if len(pk) != schnorr.PubKeyBytesLen {
		err = errorf.E("invalid pubkey len %d, require %d",
			len(pk), schnorr.PubKeyBytesLen)
		return
	}
	var p *pubkey.T
	if p, err = pubkey.New(pk); chk.E(err) {
		return
	}
	idx := prefixes.PubkeyIndex.Key(p)
	var found bo
	err = r.View(func(txn *badger.Txn) (err er) {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		// there can only be one
		for it.Rewind(); it.ValidForPrefix(idx); {
			// we already have it
			idx = it.Item().Key()
			if len(idx) != prefixes.KeySizes[prefixes.PubkeyIndex] {
				log.E.F("invalid pubkey index length got %d require %d",
					len(idx), prefixes.KeySizes[prefixes.PubkeyIndex])
				return
			}
			s := idx[len(idx)-serial.Len:]
			ser = binary.BigEndian.Uint64(s)
			found = true
			return
		}
		return
	})
	if found {
		return
	}
	// make write the new index
	ser = r.PubkeySerial()
	idx = prefixes.PubkeyIndex.Key(p, serial.Make(ser))
	err = r.Update(func(txn *badger.Txn) (err er) { return txn.Set(idx, nil) })
	return
}

// GetIndexPubkey searches for the pubkey by index, returns error if not found.
func (r *T) GetIndexPubkey(ser by) (pk by, err er) {
	k := prefixes.PubkeyIndex.Key()
	var s *serial.T
	err = r.View(func(txn *badger.Txn) (err er) {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(k); it.Next() {
			key := it.Item().Key()
			if len(key) == prefixes.KeySizes[prefixes.PubkeyIndex] {
				if s, err = serial.FromKey(key); chk.E(err) {
					continue
				}
				if equals(ser, s.Val) {
					// found it
					pk = key[1 : 1+schnorr.PubKeyBytesLen]
					return
				}
			}
		}
		return
	})
	if ser == nil {
		err = errorf.D("pubkey index for %0x not found", pk)
	}
	return
}
