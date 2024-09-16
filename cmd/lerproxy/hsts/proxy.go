package hsts

import "net/http"

type Proxy struct {
	http.Handler
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().
		Set("Strict-Transport-Security",
			"max-age=31536000; includeSubDomains; preload")
	p.ServeHTTP(w, r)
}
