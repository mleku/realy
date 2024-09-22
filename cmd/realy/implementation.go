package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/store"
)

type Relay struct {
	RatelDbPath S `envconfig:"DATABASE"`
	storage     store.I
}

func (r *Relay) Name() S {
	return "REALY"
}

func (r *Relay) Storage(ctx context.T) store.I {
	return r.storage
}

func (r *Relay) Init(path S) E {
	err := envconfig.Process("", r)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}
	return nil
}

func (r *Relay) AcceptEvent(c context.T, evt *event.T) bool {
	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 10000 {
		return false
	}

	return true
}

func (r *Relay) ServiceUrl(req *http.Request) (s S) {
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Host
	}
	proto := req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if host == "localhost" {
			proto = "ws"
		} else if strings.Index(host, ":") != -1 {
			// has a port number
			proto = "ws"
		} else if _, err := strconv.Atoi(strings.ReplaceAll(host, ".",
			"")); chk.E(err) {
			// it's a naked IP
			proto = "ws"
		} else {
			proto = "wss"
		}
	}
	return proto + "://" + host
}
