package options

import (
	"realy.lol/event"
)

type T struct {
	SkipEventFunc func(*event.T) bool
}

type O func(*T)

func Default() *T {
	return &T{}
}

func WithSkipEventFunc(skipEventFunc func(*event.T) bool) O {
	return func(o *T) {
		o.SkipEventFunc = skipEventFunc
	}
}
