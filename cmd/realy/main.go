package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-simpler.org/env"
	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/lol"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/units"
)

const AppName = "realy"

func main() {
	var err E
	var cfg *app.Config
	var help bool
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		switch arg {

		case "help", "-h", "--h", "-help", "--help", "?":
			help = true
		}
	}
	if cfg, err = app.NewConfig(); err != nil || help {
		// log.I.S(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		}
		fmt.Fprintf(os.Stderr, "Environment variables that configure %s:\n\n", AppName)
		env.Usage(cfg, os.Stderr, nil)
		fmt.Fprintf(os.Stderr, "\nCLI parameter 'help' also prints this information\n")
		fmt.Fprintf(os.Stderr,
			"\n.env file found at the ROOT_DIR/PROFILE path will be automatically loaded for configuration; set these two variables for a custom load path\n")
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
