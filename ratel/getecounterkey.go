package ratel

import (
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
)

// GetCounterKey returns the proper counter key for a given event ID.
func GetCounterKey(ser *serial.T) (key by) {
	key = prefixes.Counter.Key(ser)
	// log.T.F("counter key %d %d", index.Counter, ser.Uint64())
	return
}
