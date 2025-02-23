package realy

import (
	"net/http"
)

type Handler struct {
	http.ResponseWriter
	*http.Request
}
type Paths map[string]func(h Handler)

func Route(h Handler, p Paths) {
	for path, fn := range p {
		if path == h.URL.Path {
			fn(h)
			return
		}
	}
	// if there is a default empty string and no path matched, run the default
	if fn, ok := p[""]; ok {
		fn(h)
	}
}
