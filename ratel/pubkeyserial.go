package ratel

import (
	"realy.lol/ratel/keys/serial"
	"encoding/binary"
)

// PubkeySerial returns the next monotonic conflict free unique serial for
// pubkeys on the database.
func (r *T) PubkeySerial() (ser uint64) {
	var err er
	if ser, err = r.pubkeySeq.Next(); chk.E(err) {
		panic(err)
	}
	return
}

// PubkeySerialBytes returns a new serial value, used to store an pubkey record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *T) PubkeySerialBytes() (ser by) { return SerialToBytes(r.PubkeySerial()) }

func SerialToBytes(ser u64) (b by) {
	b = make(by, serial.Len)
	binary.BigEndian.PutUint64(b, ser)
	return
}
