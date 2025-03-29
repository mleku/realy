// Package messages is a collection of example/common messages and
// machine-readable prefixes to use with OK and CLOSED envelopes.
package messages

import (
	"lukechampine.com/frand"
)

const (
	Duplicate   = "duplicate"
	Pow         = "pow"
	Blocked     = "blocked"
	RateLimited = "rate-limited"
	Invalid     = "invalid"
	Error       = "error"
)

var Examples = [][]byte{
	[]byte(""),
	[]byte("pow: difficulty 25>=24"),
	[]byte("duplicate: already have this event"),
	[]byte("blocked: you are banned from posting here"),
	[]byte("blocked: please register your pubkey at " +
		"https://my-expensive-relay.example.com"),
	[]byte("rate-limited: slow down there chief"),
	[]byte("invalid: event creation date is too far off from the current time"),
	[]byte("pow: difficulty 26 is less than 30"),
	[]byte("error: could not connect to the database"),
}

func RandomMessage() []byte {
	return Examples[frand.Intn(len(Examples)-1)]
}
