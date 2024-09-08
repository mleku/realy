package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/fiatjaf/eventstore"
	"github.com/fiatjaf/eventstore/badger"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr"
	. "nostr.mleku.dev"
	realy "realy.mleku.dev"
)

type Relay struct {
	BadgerDbPath S `envconfig:"BADGER_DATABASE"`

	storage *badger.BadgerBackend
}

func (r *Relay) Name() S {
	return "BasicRelay"
}

func (r *Relay) Storage(ctx context.Context) eventstore.Store {
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
	r := Relay{BadgerDbPath: path}
	// if an environment variable for the database path is set, it will override the temporary.
	if err = envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	r.storage = &badger.BadgerBackend{Path: r.BadgerDbPath}
	// if err := r.storage.Init(); err != nil {
	// 	panic(err)
	// }
	var server *realy.Server
	server, err = realy.NewServer(&r)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	if err := server.Start("0.0.0.0", 3334); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
