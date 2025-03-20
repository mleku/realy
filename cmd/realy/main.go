package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/pkg/profile"

	realy_lol "realy.lol"
	"realy.lol/bech32encoding"
	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/interrupt"
	"realy.lol/lol"
	"realy.lol/p256k"
	"realy.lol/ratel"
	"realy.lol/realy"
	"realy.lol/realy/config"
	"realy.lol/realy/options"
	"realy.lol/signer"
	"realy.lol/units"
)

func main() {
	log.I.F("starting realy %s", realy_lol.Version)
	var err error
	var cfg *config.C
	if cfg, err = config.New(); chk.T(err) {
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
	if config.HelpRequested() {
		config.PrintHelp(cfg, os.Stderr)
		os.Exit(0)
	}
	log.I.Ln("log level", cfg.LogLevel)
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
			MaxLimit:       ratel.DefaultMaxLimit,
			UseCompact:     cfg.UseCompact,
			Compression:    cfg.Compression,
			Extra: []int{
				cfg.DBSizeLimit,
				cfg.DBLowWater,
				cfg.DBHighWater,
				cfg.GCFrequency * int(time.Second),
			},
		},
	)
	r := &app.Relay{C: cfg, Store: storage}
	go app.MonitorResources(c)
	var server *realy.Server
	admins := strings.Split(cfg.AdminNpubs, ",")
	var administrators []signer.I
	for _, src := range admins {
		if len(src) < 1 {
			continue
		}
		dst := make([]byte, len(src)/2)
		if _, err = hex.DecBytes(dst, []byte(src)); chk.E(err) {
			if dst, err = bech32encoding.NpubToBytes([]byte(src)); chk.E(err) {
				continue
			}
		}
		log.T.S(dst)
		sign := &p256k.Signer{}
		if err = sign.InitPub(dst); chk.E(err) {
			return
		}
		administrators = append(administrators, sign)
		log.I.F("administrator pubkey: %0x", sign.Pub())
	}
	serverParams := &realy.ServerParams{
		Ctx:            c,
		Cancel:         cancel,
		Rl:             r,
		DbPath:         cfg.DataDir,
		MaxLimit:       ratel.DefaultMaxLimit,
		Admins:         administrators,
		PublicReadable: cfg.PublicReadable,
	}
	var opts []options.O
	if len(cfg.Owners) > 0 || cfg.AuthRequired {
		log.I.F("relay requires auth for writing")
	}
	if cfg.PublicReadable {
		log.I.F("relay is public readable")
	}
	if server, err = realy.NewServer(serverParams, opts...); chk.E(err) {
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
