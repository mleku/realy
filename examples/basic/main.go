package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr"
	. "nostr.mleku.dev"
	realy "realy.mleku.dev"
	eventstore "store.mleku.dev"
	"store.mleku.dev/ratel"
	"util.mleku.dev/lol"
	"util.mleku.dev/units"
)

type Relay struct {
	RatelDbPath S `envconfig:"BADGER_DATABASE"`

	storage eventstore.I
}

func (r *Relay) Name() S {
	return "BasicRelay"
}

func (r *Relay) Storage(ctx context.Context) eventstore.I {
	return r.storage
}

func (r *Relay) Init() E {
	err := envconfig.Process("", r)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}
	return nil
}

func (r *Relay) AcceptEvent(ctx context.Context, evt *nostr.Event) bool {
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
	// default to creating a one-time temporary database
	if path, err = os.MkdirTemp("/tmp", "realy"); Chk.E(err) {
		os.Exit(1)
	}
	r := &Relay{RatelDbPath: path}
	// if an environment variable for the database path is set, it will override the temporary.
	if err = envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	var wg sync.WaitGroup
	c, cancel := context.WithCancel(context.Background())
	r.storage = ratel.GetBackend(c, &wg, r.RatelDbPath, false, units.Gb*8, lol.Trace, 0)
	var server *realy.Server
	if server, err = realy.NewServer(r); Chk.E(err) {

	}
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	if err = server.Start("0.0.0.0", 3334); Chk.E(err) {
		log.Fatalf("server terminated: %v", err)
	}
	cancel()
}
