package api

import (
	"net/http"
	"strings"
	"sync"
)

type Registry map[string]Method

var registry = make(map[string]Method)

var registryMx sync.Mutex

// RegisterCapability stores a string that describes a given method in the
// simplified nostr API.
func RegisterCapability(m Method) {
	registryMx.Lock()
	defer registryMx.Unlock()
	log.I.F("registering method for path %s", m.Path())
	registry[m.Path()] = m
}

// GetCapability returns an existing capability if it exists.
func GetCapability(c string) (m Method, ok bool) {
	registryMx.Lock()
	defer registryMx.Unlock()
	m, ok = registry[c]
	return
}

// GetCapabilities returns a new map that is a copy of the registry.
func GetCapabilities() (c map[string]Method) {
	registryMx.Lock()
	defer registryMx.Unlock()
	c = make(map[string]Method)
	for path, s := range registry {
		c[path] = s
	}
	return
}

type H struct {
	http.ResponseWriter
	*http.Request
}

func (h H) RealRemote() (rr string) {
	// reverse proxy should populate this field so we see the remote not the proxy
	rem := h.Request.Header.Get("X-Forwarded-For")
	if rem != "" {
		splitted := strings.Split(rem, " ")
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
		// in case upstream doesn't set this, or we are directly listening instead of
		// via reverse proxy or just if the header field is missing, put the
		// connection remote address into the websocket state data.
		if rr == "" {
			rr = h.Request.RemoteAddr
		}
	} else {
		// if that fails, fall back to the remote (probably the proxy, unless the realy is
		// actually directly listening)
		rr = h.Request.Host
	}
	return
}
