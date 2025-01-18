package materia

import (
	"realy.lol/ratel"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"encoding/binary"
)

// SerialKey returns a key used for storing events, and the raw serial counter
// bytes to copy into index keys.
func (r *ratel.T) SerialKey() (idx ratel.by, ser *serial.T) {
	var err ratel.er
	var s ratel.by
	if s, err = r.SerialBytes(); ratel.chk.E(err) {
		panic(err)
	}
	ser = serial.New(s)
	return prefixes.Event.Key(ser), ser
}

// Serial returns the next monotonic conflict free unique serial on the database.
func (r *ratel.T) Serial() (ser uint64, err ratel.er) {
	if ser, err = r.eventSeq.Next(); ratel.chk.E(err) {
	}
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *ratel.T) SerialBytes() (ser ratel.by, err ratel.er) {
	var serU64 uint64
	if serU64, err = r.Serial(); ratel.chk.E(err) {
		panic(err)
	}
	ser = make(ratel.by, serial.Len)
	binary.BigEndian.PutUint64(ser, serU64)
	return
}
