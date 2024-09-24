package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/lol"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/units"
)

func main() {
	var err E
	var cfg *app.Config
	if cfg, err = app.NewConfig(); err != nil || app.HelpRequested() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		}
		app.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	lol.SetLogLevel(cfg.LogLevel)
	log.T.S(cfg)
	var wg sync.WaitGroup
	c, cancel := context.Cancel(context.Bg())
	path := filepath.Join(cfg.Root, cfg.Profile)
	storage := ratel.GetBackend(c, &wg, false, units.Gb*8, lol.Trace, 0)
	r := &app.Relay{Config: cfg, Store: storage}
	var server *realy.Server
	if server, err = realy.NewServer(r, path); chk.E(err) {
		return
	}
	if err != nil {
		log.F.F("failed to create server: %v", err)
	}
	if err = server.Start(cfg.Listen, cfg.Port); chk.E(err) {
		log.F.F("server terminated: %v", err)
	}
	cancel()
}
