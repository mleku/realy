package app

import (
	"net/http"
	"strconv"
	"strings"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/store"
)

type Relay struct {
	*Config
	Store store.I
}

func (r *Relay) Name() S                                    { return "REALY" }
func (r *Relay) Storage(c context.T) store.I                { return r.Store }
func (r *Relay) Init() E                                    { return nil }
func (r *Relay) AcceptEvent(c context.T, evt *event.T) bool { return true }

func (r *Relay) ServiceUrl(req *http.Request) (s S) {
	log.I.S(req)
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
