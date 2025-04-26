// Package options provides some option configurations for the realy relay.
//
// None of this package is actually in use, and the skip event function has not been
// implemented. In theory this could be used for something but it currently isn't.
package options

import (
	"realy.lol/event"
)

type SkipEventFunc func(*event.T) bool

// T is a collection of options.
type T struct {
	// SkipEventFunc is in theory a function to test whether an event should not be sent in
	// response to a query.
	SkipEventFunc
}

// O is a function that processes an options.T.
type O func(*T)

// Default returns an uninitialised options.T.
func Default() *T {
	return &T{}
}

// WithSkipEventFunc is an options.T generator that adds a function to skip events.
func WithSkipEventFunc(skipEventFunc func(*event.T) bool) O {
	return func(o *T) {
		o.SkipEventFunc = skipEventFunc
	}
}
