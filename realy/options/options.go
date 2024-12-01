package options

import (
	"golang.org/x/time/rate"

	"realy.lol/event"
)

type T struct {
	PerConnectionLimiter *rate.Limiter
	SkipEventFunc        func(*event.T) bo
}

type O func(*T)

func Default() *T {
	return &T{}
}

func WithPerConnectionLimiter(rps rate.Limit, burst no) O {
	return func(o *T) {
		o.PerConnectionLimiter = rate.NewLimiter(rps, burst)
	}
}

func WithSkipEventFunc(skipEventFunc func(*event.T) bo) O {
	return func(o *T) {
		o.SkipEventFunc = skipEventFunc
	}
}
