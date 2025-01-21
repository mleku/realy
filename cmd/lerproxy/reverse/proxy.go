package reverse

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewSingleHostReverseProxy is a copy of httputil.NewSingleHostReverseProxy
// with addition of "X-Forwarded-Proto" header.
func NewSingleHostReverseProxy(target *url.URL) (rp *httputil.ReverseProxy) {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		log.I.S(target.Path, req.URL.Path)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		// req.URL.Path = util.SingleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("X-Forwarded-Proto", "https")
	}
	rp = &httputil.ReverseProxy{Director: director}
	return
}
