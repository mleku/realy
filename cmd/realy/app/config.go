package app

import (
	"os"
	"strings"

	"go-simpler.org/env"
	"realy.lol/apputil"
)

type Config struct {
	Root     string `env:"ROOT_DIR" usage:"root path for all other path configurations (defaults OS user home if empty)"`
	Profile  string `env:"PROFILE" default:".realy" usage:"name of directory in root path to store relay state data and database"`
	Listen   string `env:"LISTEN" default:"0.0.0.0" usage:"network listen address"`
	Port     int    `env:"PORT" default:"3334" usage:"port to listen on"`
	LogLevel string `env:"LOGLEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
}

func NewConfig() (cfg *Config, err E) {
	cfg = &Config{}
	if err = env.Load(cfg, nil); err != nil {
		return
	}
	if cfg.Root == "" {
		var dir string
		if dir, err = os.UserHomeDir(); err != nil {
			return
		}
		cfg.Root = dir
	}
	var envPath S
	if apputil.FileExists(envPath) {
		var e Env
		if e, err = GetEnv(envPath); err != nil {
			return
		}
		if err = env.Load(cfg, &env.Options{Source: e}); chk.E(err) {
			return
		}
	}
	return
}

type Env map[string]string

func GetEnv(path string) (env Env, err error) {
	var s B
	if s, err = os.ReadFile(path); err != nil {
		return
	}
	lines := strings.Split(S(s), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		split := strings.Split(line, "=")
		if len(split) != 2 {
			log.E.F("invalid line %d in config %s:\n%s", i, path, line)
			continue
		}
		env[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
	}
	return
}

func (env Env) LookupEnv(key string) (value string, ok bool) {
	value, ok = env[key]
	return
}
