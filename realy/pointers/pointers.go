package pointers

import (
	"time"

	"realy.lol/unix"
)

type PointerToValue interface {
	~*uint | ~*int | ~*uint8 | ~*uint16 | ~*uint32 | ~*uint64 | ~*int8 | ~*int16 | ~*int32 |
		~*int64 | ~*float32 | ~*float64 | ~*string | ~*[]string | ~*time.Time | ~*time.Duration |
		~*[]byte | ~*[][]byte | ~*unix.Time
}

func Present[V PointerToValue](i V) bool {
	return i != nil
}
