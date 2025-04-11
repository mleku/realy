package helpers

import (
	"net/http"
	"strings"
)

func GenerateDescription(text string, scopes []string) string {
	if len(scopes) == 0 {
		return text
	}
	result := make([]string, 0)
	for _, value := range scopes {
		result = append(result, "`"+value+"`")
	}
	return text + "<br/><br/>**Scopes**<br/>" + strings.Join(result, ", ")
}

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
