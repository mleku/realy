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
	"realy.lol/realy/config"
	"realy.lol/units"
	"realy.lol/bech32encoding"
	"realy.lol/hex"
	"realy.lol/ec/secp256k1"
	"realy.lol/p256k"
)

func main() {
	var err er
	var cfg *config.C
	if cfg, err = config.New(); chk.T(err) || config.HelpRequested() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		}
		config.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	if config.GetEnv() {
		config.PrintEnv(cfg, os.Stdout)
		os.Exit(0)
	}
	log.I.Ln("log level", cfg.LogLevel)
	var prf, spiderKey by
	var val any
	var sign *p256k.Signer
	if len(cfg.SpiderKey) > 60 {
		spiderKey = make(by, secp256k1.SecKeyBytesLen)
		log.I.F("%s", cfg.SpiderKey)
		if prf, val, err = bech32encoding.Decode(by(cfg.SpiderKey)); chk.E(err) {
			log.E.F("SPIDER_KEY decode error: '%s' hrp: %s", err.Error(), prf)
			spiderKey = nil
		} else {
			if sk, ok := val.(by); ok {
				if spiderKey, err = hex.Dec(st(sk)); chk.E(err) {
					log.E.F("failed to decode hex: '%s'", err.Error())
					spiderKey = nil
				}
				sign = &p256k.Signer{}
				if err = sign.InitSec(spiderKey); chk.E(err) {
					// if this didn't work, disable the spider key
					cfg.SpiderKey = ""
					// also nil the signer so it doesn't panic.
					sign = nil
				}
				log.I.F("signer enabled for %x", sign.Pub())
			}
		}
	}
	lol.SetLogLevel(cfg.LogLevel)
	if cfg.Pprof {
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			chk.E(http.ListenAndServe("127.0.0.1:6060", nil))
		}()
	}
	debug.SetMemoryLimit(int64(cfg.MemLimit))
	var wg sync.WaitGroup
	c, cancel := context.Cancel(context.Bg())
	storage := ratel.New(
		ratel.BackendParams{
			Ctx:            c,
			WG:             &wg,
			BlockCacheSize: units.Gb,
			LogLevel:       lol.GetLogLevel(cfg.DbLogLevel),
			MaxLimit:       cfg.MaxLimit,
			UseCompact:     cfg.UseCompact,
			Compression:    cfg.Compression,
			Extra: []no{
				cfg.DBSizeLimit,
				cfg.DBLowWater,
				cfg.DBHighWater,
				cfg.GCFrequency * no(time.Second),
			},
		},
	)
	r := &app.Relay{
		Ctx:          c,
		C:            cfg,
		Store:        storage,
		MaxLimit:     cfg.MaxLimit,
		SpiderSigner: sign,
	}
	go app.MonitorResources(c)
	var server *realy.Server
	if server, err = realy.NewServer(realy.ServerParams{
		Ctx:       c,
		Cancel:    cancel,
		Rl:        r,
		DbPath:    cfg.Profile,
		MaxLimit:  cfg.MaxLimit,
		AdminUser: cfg.AdminUser,
		AdminPass: cfg.AdminPass,
	}); chk.E(err) {
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
