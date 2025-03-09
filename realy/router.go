package realy

import (
	"net/http"
)

type Handler struct {
	http.ResponseWriter
	*http.Request
}

type Protocol map[string]func(h Handler)

type Paths map[string]Protocol

func Route(h Handler, p Paths) {
	acc := h.Request.Header.Get("Accept")
	log.I.S(acc)
	for proto, fns := range p {
		if proto == acc {
			for path, fn := range fns {
				if path == h.URL.Path {
					fn(h)
					return
				}
			}
		}
	}
}
