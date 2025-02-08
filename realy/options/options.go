package options

import (
	"golang.org/x/time/rate"

	"realy.lol/event"
)

type T struct {
	PerConnectionLimiter *rate.Limiter
	SkipEventFunc        func(*event.T) bool
}

type O func(*T)

func Default() *T {
	return &T{}
}

func WithPerConnectionLimiter(rps rate.Limit, burst int) O {
	return func(o *T) {
		o.PerConnectionLimiter = rate.NewLimiter(rps, burst)
	}
}

func WithSkipEventFunc(skipEventFunc func(*event.T) bool) O {
	return func(o *T) {
		o.SkipEventFunc = skipEventFunc
	}
}
