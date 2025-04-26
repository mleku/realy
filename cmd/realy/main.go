// Package main is a nostr relay with a simple follow/mute list authentication
// scheme and the new HTTP REST based protocol. Configuration is via environment
// variables or an optional .env file.
package main

import (
	"net"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/adrg/xdg"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/chk"
	"realy.lol/config"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/interrupt"
	"realy.lol/log"
	"realy.lol/lol"
	"realy.lol/openapi"
	"realy.lol/p256k"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/servemux"
	"realy.lol/socketapi"
	"realy.lol/units"
)

func main() {
	log.I.F("starting realy %s", realy_lol.Version)
	cfg := config.New()
	if cfg.Superuser == "" {
		log.F.F("SUPERUSER is not set")
		os.Exit(1)
	}
	runtime.GOMAXPROCS(1)
	debug.SetGCPercent(10)
	debug.SetMemoryLimit(250000)
	var err error
	a := cfg.Superuser
	var dst []byte
	if dst, err = bech32encoding.NpubToBytes([]byte(a)); chk.E(err) {
		if _, err = hex.DecBytes(dst, []byte(a)); chk.E(err) {
			log.F.F("SUPERUSER is invalid: %s", a)
			os.Exit(1)
		}
	}
	super := &p256k.Signer{}
	if err = super.InitPub(dst); chk.E(err) {
		return
	}
	lol.ShortLoc.Store(false)
	log.I.F("starting %s %s", cfg.AppName, realy_lol.Version)
	wg := &sync.WaitGroup{}
	c, cancel := context.Cancel(context.Bg())
	interrupt.AddHandler(func() { cancel() })
	storage := ratel.New(
		ratel.BackendParams{
			Ctx:            c,
			WG:             wg,
			BlockCacheSize: units.Gb,
			LogLevel:       lol.Info,
			MaxLimit:       ratel.DefaultMaxLimit,
		},
	)
	if err = storage.Init(filepath.Join(xdg.DataHome, cfg.AppName)); chk.E(err) {
		os.Exit(1)
	}
	serveMux := servemux.New()
	s := &realy.Server{
		Name:      cfg.AppName,
		Ctx:       c,
		Cancel:    cancel,
		WG:        wg,
		Mux:       serveMux,
		Address:   net.JoinHostPort(cfg.Listen, strconv.Itoa(cfg.Port)),
		Store:     storage,
		MaxLimit:  ratel.DefaultMaxLimit,
		Superuser: super,
	}
	openapi.New(s, cfg.AppName, realy_lol.Version, realy_lol.Description, "/api", serveMux)
	socketapi.New(s, "/{$}", serveMux)
	interrupt.AddHandler(func() { s.Shutdown() })
	if err = s.Start(); chk.E(err) {
		os.Exit(1)
	}
}
