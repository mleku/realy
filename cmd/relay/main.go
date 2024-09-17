package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"realy.lol/event"
	"realy.lol/lol"
	"realy.lol/ratel"
	realy "realy.lol/relay"
	"realy.lol/store"
	"realy.lol/units"
)

type Relay struct {
	RatelDbPath S `envconfig:"DATABASE"`
	storage     store.I
}

func (r *Relay) Name() S {
	return "REALY"
}

func (r *Relay) Storage(ctx context.Context) store.I {
	return r.storage
}

func (r *Relay) Init(path S) E {
	err := envconfig.Process("", r)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}
	return nil
}

func (r *Relay) AcceptEvent(ctx context.Context, evt *event.T) bool {
	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 10000 {
		return false
	}

	return true
}

func main() {
	var err E
	var path S
	r := &Relay{}
	if err = envconfig.Process(r.Name(), r); err != nil {
		log.F.F("failed to read from env: %v", err)
		return
	}
	log.I.F("'%s'", r.RatelDbPath)
	if r.RatelDbPath == "" {
		// default to creating a one-time temporary database
		if path, err = os.MkdirTemp("/tmp", "realy"); chk.E(err) {
			os.Exit(1)
		}
		r.RatelDbPath = path
	}
	log.I.F("'%s'", r.RatelDbPath)
	var wg sync.WaitGroup
	c, cancel := context.WithCancel(context.Background())
	r.storage = ratel.GetBackend(c, &wg, r.RatelDbPath, false, units.Gb*8,
		lol.Trace, 0)
	var server *realy.Server
	if server, err = realy.NewServer(r, path); chk.E(err) {
		return
	}
	if err != nil {
		log.F.F("failed to create server: %v", err)
	}
	if err = server.Start("0.0.0.0", 3334); chk.E(err) {
		log.F.F("server terminated: %v", err)
	}
	cancel()
}
