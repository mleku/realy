// Package main is a nostr relay with a simple follow/mute list authentication
// scheme and the new HTTP REST based protocol. Configuration is via environment
// variables or an optional .env file.
package main

import (
	"errors"
	"net"
	"net/http/httputil"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/adrg/xdg"

	realy_lol "realy.mleku.dev"
	"realy.mleku.dev/chk"
	"realy.mleku.dev/config"
	"realy.mleku.dev/context"
	"realy.mleku.dev/interrupt"
	"realy.mleku.dev/log"
	"realy.mleku.dev/lol"
	"realy.mleku.dev/openapi"
	"realy.mleku.dev/ratel"
	"realy.mleku.dev/realy"
	"realy.mleku.dev/servemux"
	"realy.mleku.dev/socketapi"
	"realy.mleku.dev/units"
)

func main() {
	log.I.F("starting realy %s", realy_lol.Version)
	cfg := config.New()
	lol.ShortLoc.Store(false)
	log.I.F("starting %s %s", cfg.AppName, realy_lol.Version)
	wg := &sync.WaitGroup{}
	c, cancel := context.Cancel(context.Bg())
	interrupt.AddHandler(func() { cancel() })
	storage := ratel.New(
		ratel.BackendParams{
			Ctx:            c,
			WG:             wg,
			BlockCacheSize: 250 * units.Mb,
			LogLevel:       lol.Info,
			MaxLimit:       ratel.DefaultMaxLimit,
			UseCompact:     false,
			Compression:    "zstd",
		},
	)
	var err error
	if err = storage.Init(filepath.Join(xdg.DataHome, cfg.AppName)); chk.E(err) {
		os.Exit(1)
	}
	serveMux := servemux.New()
	s := &realy.Server{
		Name:     cfg.AppName,
		Ctx:      c,
		Cancel:   cancel,
		WG:       wg,
		Mux:      serveMux,
		Address:  net.JoinHostPort(cfg.Listen, strconv.Itoa(cfg.Port)),
		Store:    storage,
		MaxLimit: ratel.DefaultMaxLimit,
	}
	openapi.New(s, cfg.AppName, realy_lol.Version, realy_lol.Description, "/api", serveMux)
	socketapi.New(s, "/{$}", serveMux)
	interrupt.AddHandler(func() { s.Shutdown() })
	if err = s.Start(); err != nil {
		if errors.Is(err, httputil.ErrClosed) {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
