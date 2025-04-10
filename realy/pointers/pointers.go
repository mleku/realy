package pointers

import (
	"time"

	"realy.mleku.dev/unix"
)

// PointerToValue is a generic interface to refer to any pointer to almost any kind of common
// type of value.
type PointerToValue interface {
	~*uint | ~*int | ~*uint8 | ~*uint16 | ~*uint32 | ~*uint64 | ~*int8 | ~*int16 | ~*int32 |
		~*int64 | ~*float32 | ~*float64 | ~*string | ~*[]string | ~*time.Time | ~*time.Duration |
		~*[]byte | ~*[][]byte | ~*unix.Time
}

// Present determines whether there is a value for a PointerToValue type.
func Present[V PointerToValue](i V) bool {
	return i != nil
}
