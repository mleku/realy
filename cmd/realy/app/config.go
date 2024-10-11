package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"go-simpler.org/env"
	"realy.lol/apputil"
	"realy.lol/config"
)

type Config struct {
	AppName      S    `env:"APP_NAME" default:"realy"`
	Root         S    `env:"ROOT_DIR" usage:"root path for all other path configurations (defaults OS user home if empty)"`
	Profile      S    `env:"PROFILE" default:".realy" usage:"name of directory in root path to store relay state data and database"`
	Listen       S    `env:"LISTEN" default:"0.0.0.0" usage:"network listen address"`
	Port         N    `env:"PORT" default:"3334" usage:"port to listen on"`
	AdminListen  S    `env:"ADMIN_LISTEN" default:"127.0.0.1" usage:"admin listen address"`
	AdminPort    N    `env:"ADMIN_PORT" default:"3337" usage:"admin listen port"`
	LogLevel     S    `env:"LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	DbLogLevel   S    `env:"DB_LOG_LEVEL" default:"info" usage:"debug level: fatal error warn info debug trace"`
	AuthRequired bool `env:"AUTH_REQUIRED" default:"false" usage:"requires auth for all access"`
	Owners       []S  `env:"OWNERS" usage:"list of npubs of users in hex format whose follow and mute list dictate accepting requests and events - follows and follows follows are allowed, mutes and follows mutes are rejected"`
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
		// if err = env.Load(cfg, nil); err != nil {
		// 	return
		// }
	}
	return
}

// HelpRequested returns true if any of the common types of help invocation are
// found as the first command line parameter/flag.
func HelpRequested() (help bool) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "help", "-h", "--h", "-help", "--help", "?":
			help = true
		}
	}
	return
}

func GetEnv() (requested bool) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "env":
			requested = true
		}
	}
	return
}

func PrintEnv(cfg *Config, printer io.Writer) {
	t := reflect.TypeOf(*cfg)

	for i := 0; i < t.NumField(); i++ {
		k := t.Field(i).Tag.Get("env")
		v := reflect.ValueOf(*cfg).Field(i).Interface()
		var val S
		switch v.(type) {
		case string:
			val = v.(string)
		case int, bool:
			val = fmt.Sprint(v)
		case []string:
			arr := v.([]string)
			if len(arr) > 0 {
				val = strings.Join(arr, ",")
			}
		}
		fmt.Fprintf(printer, "%s=%v\n", k, val)
	}
}

// PrintHelp outputs a help text listing the configuration options and default
// values to a provided io.Writer (usually os.Stderr or os.Stdout).
func PrintHelp(cfg *Config, printer io.Writer) {
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
			"set the environment using\n\n\t%s env>%s/%s/.env\n\n", os.Args[0], cfg.Root,
		cfg.Profile)
	return
}
