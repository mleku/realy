package realy

import (
	"net/http"
	"strings"
)

func GetRemoteFromReq(r *http.Request) (rr string) {
	// reverse proxy should populate this field so we see the remote not the proxy
	rr = r.Header.Get("X-Forwarded-For")
	if rr != "" {
		splitted := strings.Split(rr, " ")
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
		// in case upstream doesn't set this or we are directly listening instead of
		// via reverse proxy or just if the header field is missing, put the
		// connection remote address into the websocket state data.
		if rr == "" {
			rr = r.RemoteAddr
		}
	}
	return
}
