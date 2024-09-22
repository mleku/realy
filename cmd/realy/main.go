package main

import (
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"realy.lol/context"
	"realy.lol/lol"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/units"
)

func main() {
	lol.SetLogLevel("debug")
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
	c, cancel := context.Cancel(context.Bg())
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
