package options

import (
	"golang.org/x/time/rate"

	"realy.lol/event"
)

type O struct {
	PerConnectionLimiter *rate.Limiter
	SkipEventFunc        func(*event.T) bo
}

type F func(*O)

func Default() *O {
	return &O{}
}

func WithPerConnectionLimiter(rps rate.Limit, burst no) F {
	return func(o *O) {
		o.PerConnectionLimiter = rate.NewLimiter(rps, burst)
	}
}

func WithSkipEventFunc(skipEventFunc func(*event.T) bo) F {
	return func(o *O) {
		o.SkipEventFunc = skipEventFunc
	}
}
