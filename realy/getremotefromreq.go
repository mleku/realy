package realy

import (
	"net/http"
	"strings"
)

func GetRemoteFromReq(r *http.Request) (rr string) {
	// reverse proxy should populate this field so we see the remote not the proxy
	rem := r.Header.Get("X-Forwarded-For")
	if rem == "" {
		rr = r.RemoteAddr
	} else {
		splitted := strings.Split(rem, " ")
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
	}
	return
}
