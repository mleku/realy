package config

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/profile"
	"go-simpler.org/env"

	"realy.lol"
	"realy.lol/chk"
	"realy.lol/config/keyvalue"
)

// C is the configuration for a realy. Note that it is absolutely minimal. More complex
// configurations should generally be stored in the database, where APIs make them easy to
// modify.
type C struct {
	AppName   string `env:"APP_NAME" default:"realy"`
	Listen    string `env:"LISTEN" default:"0.0.0.0" usage:"network listen address"`
	Port      int    `env:"PORT" default:"3334" usage:"network listen port"`
	Pprof     bool   `env:"PPROF" default:"false" usage:"enable pprof on 127.0.0.1:6060"`
	Superuser string `env:"SUPERUSER" usage:"superuser npub/hex public key"`
	Binary    bool   `env:"BINARY" usage:"use binary encoder for database" default:"false"`
}

func New() (c *C) {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(realy_lol.Version)
		os.Exit(0)
	}
	c = &C{}
	if err := env.Load(c, &env.Options{SliceSep: ","}); chk.T(err) {
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "help" {
		fmt.Printf("\nenvironment variables that configure %s\n\n", c.AppName)
		env.Usage(c, os.Stdout, nil)
		fmt.Printf(`
commands:

  - print this help message

      %s help    

  - print version info

      %s version 

  - print environment variables as a shell script that can be edited to set the configuration

      %s env 

`, os.Args[0], os.Args[0], os.Args[0])
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "env" {
		keyvalue.PrintEnv(*c, os.Stdout)
		os.Exit(0)
	}
	// now we have the config, set up all the things here rather than somewhere unrelated.
	if c.Pprof {
		defer profile.Start(profile.MemProfile).Stop()
		go func() {
			chk.E(http.ListenAndServe("127.0.0.1:6060", nil))
		}()
	}
	return
}
