package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/profile"
	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/interrupt"
	"realy.lol/lol"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/units"
)

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	var err E
	var cfg *app.Config
	if cfg, err = app.NewConfig(); chk.T(err) || app.HelpRequested() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		}
		app.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	if app.GetEnv() {
		app.PrintEnv(cfg, os.Stdout)
		os.Exit(0)
	}
	log.I.Ln("log level", cfg.LogLevel)
	lol.SetLogLevel(cfg.LogLevel)
	var wg sync.WaitGroup
	c, cancel := context.Cancel(context.Bg())
	path := filepath.Join(cfg.Root, cfg.Profile)
	storage := ratel.GetBackend(c, &wg, false, units.Gb*166, lol.GetLogLevel(cfg.DbLogLevel),
		ratel.DefaultMaxLimit)
	r := &app.Relay{Config: cfg, Store: storage}
	go app.MonitorResources(c)
	var server *realy.Server
	if server, err = realy.NewServer(c, cancel, r, path); chk.E(err) {
		os.Exit(1)
	}
	if err != nil {
		log.F.F("failed to create server: %v", err)
	}
	interrupt.AddHandler(func() { server.Shutdown() })
	if err = server.Start(cfg.Listen, cfg.Port, cfg.AdminListen, cfg.AdminPort); chk.E(err) {
		log.F.F("server terminated: %v", err)
	}
	// cancel()
}
