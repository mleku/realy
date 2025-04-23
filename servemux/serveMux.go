package servemux

import "net/http"

type S struct {
	*http.ServeMux
}

func New() *S {
	return &S{http.NewServeMux()}
}

func (c *S) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		return
	}
	c.ServeMux.ServeHTTP(w, r)
}
