package realy

import (
	"net/http"
)

type Handler struct {
	http.ResponseWriter
	*http.Request
}

type Protocol map[string]func(w http.ResponseWriter, r *http.Request)

type Paths map[string]Protocol

func Route(w http.ResponseWriter, r *http.Request, p Paths) {
	acc := r.Header.Get("Accept")
	log.I.S(acc)
	for proto, fns := range p {
		if proto == acc || proto == "" {
			for path, fn := range fns {
				if path == r.URL.Path {
					fn(w, r)
					return
				}
			}
		}
	}
}
