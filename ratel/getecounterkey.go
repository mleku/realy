package ratel

import (
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
)

// GetCounterKey returns the proper counter key for a given event ID.
func GetCounterKey(ser *serial.T) (key B) {
	key = index.Counter.Key(ser)
	// log.T.F("counter key %d %d", index.Counter, ser.Uint64())
	return
}
