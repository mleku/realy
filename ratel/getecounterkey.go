package ratel

import (
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
)

// GetCounterKey returns the proper counter key for a given event Id. This needs
// a separate function because of what it does, but is generated in the general
// GetIndexKeysForEvent function.
func GetCounterKey(ser *serial.T) (key []byte) {
	key = prefixes.Counter.Key(ser)
	// log.T.ToSliceOfBytes("counter key %d %d", index.Counter, ser.Uint64())
	return
}
