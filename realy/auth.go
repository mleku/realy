package realy

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"realy.lol/chk"
	"realy.lol/httpauth"
	"realy.lol/log"
)

func (s *Server) adminAuth(r *http.Request,
	tolerance ...time.Duration) (authed bool, pubkey []byte) {
	var valid bool
	var err error
	var tolerate time.Duration
	if len(tolerance) > 0 {
		tolerate = tolerance[0]
	}
	if valid, pubkey, err = httpauth.CheckAuth(r, tolerate); chk.E(err) {
		return
	}
	if !valid {
		log.E.F("invalid auth %s", r.Header.Get("Authorization"))
		return
	}
	if bytes.Equal(s.Superuser.Pub(), pubkey) {
		authed = true
		return
	}
	// check admins pubkey list
	for _, v := range s.admins {
		log.I.F("%0x", v.Pub())
		if bytes.Equal(v.Pub(), pubkey) {
			authed = true
			return
		}
	}
	return
}

// ServiceURL returns the address of the relay to send back in auth responses.
// If auth is disabled this returns an empty string.
func (s *Server) ServiceURL(req *http.Request) (st string) {
	if !s.AuthRequired() && len(s.Owners()) == 0 {
		log.T.F("auth not required")
		return
	}
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}
	proto := req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if host == "localhost" {
			proto = "ws"
		} else if strings.Contains(host, ":") {
			// has a port number
			proto = "ws"
		} else if _, err := strconv.Atoi(strings.ReplaceAll(host, ".",
			"")); chk.E(err) {
			// it's a naked IP
			proto = "ws"
		} else {
			proto = "wss"
		}
	} else if proto == "https" {
		proto = "wss"
	} else if proto == "http" {
		proto = "ws"
	}
	return proto + "://" + host
}
