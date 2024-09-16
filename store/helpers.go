package eventstore

import (
	"bytes"

	"realy.lol/event"
)

func isOlder(prev, next *event.T) bool {
	return prev.CreatedAt.I64() < next.CreatedAt.I64() ||
		(prev.CreatedAt == next.CreatedAt && bytes.Compare(prev.ID, next.ID) < 0)
}
