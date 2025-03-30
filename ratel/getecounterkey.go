package ratel

import (
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
)

// GetCounterKey returns the proper counter key for a given event Id. This needs
// a separate function because of what it does, but is generated in the general
// GetIndexKeysForEvent function.
func GetCounterKey(ser *serial.T) (key []byte) {
	key = prefixes.Counter.Key(ser)
	// log.T.F("counter key %d %d", index.Counter, ser.Uint64())
	return
}
