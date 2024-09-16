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

var Examples = []B{
	B(""),
	B("pow: difficulty 25>=24"),
	B("duplicate: already have this event"),
	B("blocked: you are banned from posting here"),
	B("blocked: please register your pubkey at " +
		"https://my-expensive-relay.example.com"),
	B("rate-limited: slow down there chief"),
	B("invalid: event creation date is too far off from the current time"),
	B("pow: difficulty 26 is less than 30"),
	B("error: could not connect to the database"),
}

func RandomMessage() B {
	return Examples[frand.Intn(len(Examples)-1)]
}
