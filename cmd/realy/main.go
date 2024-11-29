package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"sync"
	"time"

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
	if cfg.Pprof {
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			http.ListenAndServe("127.0.0.1:6060", nil)
		}()
	}
	debug.SetMemoryLimit(int64(cfg.MemLimit))
	var wg sync.WaitGroup
	c, cancel := context.Cancel(context.Bg())
	storage := ratel.New(
		ratel.BackendParams{
			Ctx:            c,
			WG:             &wg,
			BlockCacheSize: units.Gb * 16,
			LogLevel:       lol.GetLogLevel(cfg.DbLogLevel),
			MaxLimit:       ratel.DefaultMaxLimit,
			Extra: []int{
				cfg.DBSizeLimit,
				cfg.DBLowWater,
				cfg.DBHighWater,
				cfg.GCFrequency * int(time.Second),
			},
		},
	)
	r := &app.Relay{Config: cfg, Store: storage}
	go app.MonitorResources(c)
	var server *realy.Server
	if server, err = realy.NewServer(realy.ServerParams{
		Ctx:       c,
		Cancel:    cancel,
		Rl:        r,
		DbPath:    cfg.Profile,
		MaxLimit:  ratel.DefaultMaxLimit,
		AdminUser: cfg.AdminUser,
		AdminPass: cfg.AdminPass}); chk.E(err) {

		os.Exit(1)
	}
	if err != nil {
		log.F.F("failed to create server: %v", err)
	}
	interrupt.AddHandler(func() { server.Shutdown() })
	if err = server.Start(cfg.Listen, cfg.Port); chk.E(err) {
		log.F.F("server terminated: %v", err)
	}
	// cancel()
}
