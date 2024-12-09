package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"go-simpler.org/env"

	"realy.lol/appdata"
	"realy.lol/apputil"
	"realy.lol/config"
	"realy.lol/sha256"
)

type C struct {
	AppName      st   `env:"APP_NAME" default:"realy"`
	Profile      st   `env:"PROFILE" usage:"root path for all other path configurations (based on APP_NAME and OS specific location)"`
	Listen       st   `env:"LISTEN" default:"0.0.0.0" usage:"network listen address"`
	Port         no   `env:"PORT" default:"3334" usage:"port to listen on"`
	AdminUser    st   `env:"ADMIN_USER" default:"admin" usage:"admin user"`
	AdminPass    st   `env:"ADMIN_PASS" usage:"admin password"`
	LogLevel     st   `env:"LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	DbLogLevel   st   `env:"DB_LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	AuthRequired bo   `env:"AUTH_REQUIRED" default:"false" usage:"requires auth for all access"`
	Owners       []st `env:"OWNERS" usage:"list of npubs of users in hex format whose follow and mute list dictate accepting requests and events with AUTH_REQUIRED enabled - follows and follows follows are allowed to read/write, owners mutes events are rejected"`
	DBSizeLimit  no   `env:"DB_SIZE_LIMIT" default:"0" usage:"the number of gigabytes (1,000,000,000 bytes) we want to keep the data store from exceeding, 0 means disabled"`
	DBLowWater   no   `env:"DB_LOW_WATER" default:"60" usage:"the percentage of DBSizeLimit a GC run will reduce the used storage down to"`
	DBHighWater  no   `env:"DB_HIGH_WATER" default:"80" usage:"the trigger point at which a GC run should start if exceeded"`
	GCFrequency  no   `env:"GC_FREQUENCY" default:"3600" usage:"the frequency of checks of the current utilisation in minutes"`
	Pprof        bo   `env:"PPROF" default:"false" usage:"enable pprof on 127.0.0.1:6060"`
	MemLimit     no   `env:"MEMLIMIT" default:"250000000" usage:"set memory limit, default is 250Mb"`
	UseCompact   bo   `env:"USE_COMPACT" default:"false" usage:"use the compact database encoding for the ratel event store"`
	Compression  st   `env:"COMPRESSION" default:"none" usage:"compress the database, [none|snappy|zstd]"`
	// NWC          st   `env:"NWC" usage:"NWC connection string for relay to interact with an NWC enabled wallet"` // todo
}

func New() (cfg *C, err er) {
	cfg = &C{}
	if err = env.Load(cfg, nil); chk.T(err) {
		return
	}
	if cfg.Profile == "" {
		cfg.Profile = appdata.Dir(cfg.AppName, true)
	}
	envPath := filepath.Join(cfg.Profile, ".env")
	if apputil.FileExists(envPath) {
		var e config.Env
		if e, err = config.GetEnv(envPath); chk.T(err) {
			return
		}
		if err = env.Load(cfg, &env.Options{Source: e}); chk.E(err) {
			return
		}
		var owners []st
		// remove empties if any
		for _, o := range cfg.Owners {
			if len(o) == sha256.Size*2 {
				owners = append(owners, o)
			}
		}
		cfg.Owners = owners
	}
	return
}

// HelpRequested returns true if any of the common types of help invocation are
// found as the first command line parameter/flag.
func HelpRequested() (help bo) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "help", "-h", "--h", "-help", "--help", "?":
			help = true
		}
	}
	return
}

func GetEnv() (requested bo) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "env":
			requested = true
		}
	}
	return
}

type KV struct{ Key, Value st }

type KVSlice []KV

func (kv KVSlice) Len() int           { return len(kv) }
func (kv KVSlice) Less(i, j int) bool { return kv[i].Key < kv[j].Key }
func (kv KVSlice) Swap(i, j int)      { kv[i], kv[j] = kv[j], kv[i] }

// Composit merges two KVSlice together, replacing the values of earlier keys with same named
// KV items later in the slice (enabling compositing two together as a .env, as well as them
// being composed as structs.
func (kv KVSlice) Composit(kv2 KVSlice) (out KVSlice) {
	// duplicate the initial KVSlice
	for _, p := range kv {
		out = append(out, p)
	}
out:
	for i, p := range kv2 {
		for j, q := range out {
			// if the key is repeated, replace the value
			if p.Key == q.Key {
				out[j].Value = kv2[i].Value
				continue out
			}
		}
		out = append(out, p)
	}
	return
}

// EnvKV turns a struct with `env` keys (used with go-simpler/env) into a standard formatted
// environment variable key/value pair list, one per line. Note you must dereference a pointer
// type to use this. This allows the composition of the config in this file with an extended
// form with a customized variant of realy to produce correct environment variables both read
// and write.
func EnvKV(cfg any) (m KVSlice) {
	t := reflect.TypeOf(cfg)
	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Tag.Get("env")
		v := reflect.ValueOf(cfg).Field(i).Interface()
		var val st
		switch v.(type) {
		case string:
			val = v.(string)
		case no, bo, time.Duration:
			val = fmt.Sprint(v)
		case []string:
			arr := v.([]string)
			if len(arr) > 0 {
				val = strings.Join(arr, ",")
			}
		}
		// this can happen with embedded structs
		if k == "" {
			continue
		}
		m = append(m, KV{k, val})
	}
	return
}

func PrintEnv(cfg *C, printer io.Writer) {
	kvs := EnvKV(*cfg)
	sort.Sort(kvs)
	for _, v := range kvs {
		_, _ = fmt.Fprintf(printer, "%s=%s\n", v.Key, v.Value)
	}
}

// PrintHelp outputs a help text listing the configuration options and default
// values to a provided io.Writer (usually os.Stderr or os.Stdout).
func PrintHelp(cfg *C, printer io.Writer) {
	_, _ = fmt.Fprintf(printer,
		"Environment variables that configure %s:\n\n", cfg.AppName)
	env.Usage(cfg, printer, &env.Options{SliceSep: ","})
	_, _ = fmt.Fprintf(printer,
		"\nCLI parameter 'help' also prints this information\n"+
			"\n.env file found at the ROOT_DIR/PROFILE path will be automatically "+
			"loaded for configuration.\nset these two variables for a custom load path,"+
			" this file will be created on first startup.\nenvironment overrides it and "+
			"you can also edit the file to set configuration options\n\n"+
			"use the parameter 'env' to print out the current configuration to the terminal\n\n"+
			"set the environment using\n\n\t%s env>%s/%s/.env\n\n", os.Args[0], cfg.Profile,
		cfg.Profile)
	return
}
