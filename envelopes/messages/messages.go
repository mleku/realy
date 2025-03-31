// Package messages is a collection of example/common messages and
// machine-readable prefixes to use with OK and CLOSED envelopes.
package messages

import (
	"lukechampine.com/frand"
)

const (
	// Duplicate is a machine readable prefix for OK envelopes indicating that the
	// submitted event is already in the relay,s event store.
	Duplicate = "duplicate"

	// Pow is a machine readable prefix for OK envelopes indicating that the
	// eventid.T lacks sufficient zeros at the front.
	Pow = "pow"

	// Blocked is a machine readable prefix for OK envelopes indicating the event
	// submission or REQ has been rejected.
	Blocked = "blocked"

	// RateLimited is a machine readable prefix for CLOSED and OK envelopes
	// indicating the relay is now slowing down processing of requests from the
	// client.
	RateLimited = "rate-limited"

	// Invalid is a machine readable prefix for OK envelopes indicating
	// that the submitted event or other request is not correctly formatted, and may
	// mean a signature does not verify.
	Invalid = "invalid"

	// Error is a machine readable prefix for CLOSED and OK envelopes indicating
	// there was some kind of error in processing the request.
	Error = "error"
)

// Examples are some examples of the use of the prefixes above with appropriate
// human-readable suffixes.
var Examples = [][]byte{
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

// RandomMessage generates a random message out of the above list of Examples.
func RandomMessage() []byte {
	return Examples[frand.Intn(len(Examples)-1)]
}
