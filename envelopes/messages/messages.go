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

var Examples = []by{
	by(""),
	by("pow: difficulty 25>=24"),
	by("duplicate: already have this event"),
	by("blocked: you are banned from posting here"),
	by("blocked: please register your pubkey at " +
		"https://my-expensive-relay.example.com"),
	by("rate-limited: slow down there chief"),
	by("invalid: event creation date is too far off from the current time"),
	by("pow: difficulty 26 is less than 30"),
	by("error: could not connect to the database"),
}

func RandomMessage() by {
	return Examples[frand.Intn(len(Examples)-1)]
}
