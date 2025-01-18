package ratel

// // SerialKey returns a key used for storing events, and the raw serial counter
// // bytes to copy into index keys.
// func (r *T) SerialKey() (idx by, ser *serial.T) {
// 	var err er
// 	var s by
// 	if s, err = r.SerialBytes(); chk.E(err) {
// 		panic(err)
// 	}
// 	ser = serial.New(s)
// 	return prefixes.Event.Key(ser), ser
// }

// Serial returns the next monotonic conflict free unique serial on the
// database. It panics if the Next function returns an error because it
// shouldn't ever.
func (r *T) Serial() (ser uint64) {
	var err er
	if ser, err = r.eventSeq.Next(); chk.E(err) {
		panic(err)
	}
	return
}

// SerialBytes returns a new serial value, used to store an event record with a
// conflict-free unique code (it is a monotonic, atomic, ascending counter).
func (r *T) SerialBytes() (ser by) { return SerialToBytes(r.Serial()) }
