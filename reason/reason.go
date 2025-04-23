package reason

import (
	"bytes"
	"fmt"
)

// R is the machine-readable prefix before the colon in an OK or CLOSED envelope message.
// Below are the most common kinds that are mentioned in NIP-01.
type R []byte

var (
	AuthRequired = R("auth-required")
	PoW          = R("pow")
	Duplicate    = R("duplicate")
	Blocked      = R("blocked")
	RateLimited  = R("rate-limited")
	Invalid      = R("invalid")
	Error        = R("error")
	Unsupported  = R("unsupported")
	Restricted   = R("restricted")
)

// S returns the R as a string
func (r R) S() string { return string(r) }

// B returns the R as a byte slice.
func (r R) B() []byte { return r }

// IsPrefix returns whether a text contains the same R prefix.
func (r R) IsPrefix(reason []byte) bool { return bytes.HasPrefix(reason, r.B()) }

// F allows creation of a full R text with a printf style format.
func (r R) F(format string, params ...any) []byte {
	return Msg(r, format, params...)
}

// Msg constructs a properly formatted message with a machine-readable prefix for OK and CLOSED
// envelopes.
func Msg(prefix R, format string, params ...any) []byte {
	if len(prefix) < 1 {
		prefix = Error
	}
	return []byte(fmt.Sprintf(prefix.S()+": "+format, params...))
}
