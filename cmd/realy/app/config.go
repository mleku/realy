package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-simpler.org/env"
	"realy.lol/apputil"
	"realy.lol/config"
)

type Config struct {
	AppName  string `env:"APP_NAME" default:"realy"`
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
	envPath := filepath.Join(filepath.Join(cfg.Root, cfg.Profile), ".env")
	if apputil.FileExists(envPath) {
		var e config.Env
		if e, err = config.GetEnv(envPath); err != nil {
			return
		}
		if err = env.Load(cfg, &env.Options{Source: e}); chk.E(err) {
			return
		}
		// load the environment vars again so they can override the .env file
		if err = env.Load(cfg, nil); err != nil {
			return
		}
	}
	return
}

func HelpRequested() (help bool) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "help", "-h", "--h", "-help", "--help", "?":
			help = true
		}
	}
	return
}

func PrintHelp(cfg *Config, printer io.Writer) (s S) {
	_, _ = fmt.Fprintf(printer,
		"Environment variables that configure %s:\n\n", cfg.AppName)
	env.Usage(cfg, printer, nil)
	_, _ = fmt.Fprintf(printer,
		"\nCLI parameter 'help' also prints this information\n"+
			"\n.env file found at the ROOT_DIR/PROFILE path will be automatically "+
			"loaded for configuration; set these two variables for a custom load path\n")
	return
}
