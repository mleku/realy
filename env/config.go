// Package env is an implementation of the env.Source interface from
// go-simpler.org
package env

import (
	"os"
	"strings"

	"realy.mleku.dev/chk"
)

// Env is a key/value map used to represent environment variables. This is
// implemented for go-simpler.org library.
type Env map[string]string

// GetEnv reads a file expected to represent a collection of KEY=value in
// standard shell environment variable format - ie, key usually in all upper
// case no spaces and words separated by underscore, value can have any
// separator, but usually comma, for an array of values.
func GetEnv(path string) (env Env, err error) {
	var s []byte
	env = make(Env)
	if s, err = os.ReadFile(path); chk.T(err) {
		return
	}
	lines := strings.Split(string(s), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		split := strings.SplitN(line, "=", 2)
		env[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
	}
	return
}

// LookupEnv returns the raw string value associated with a provided key name,
// used as a custom environment variable loader for go-simpler.org/env to enable
// .env file loading.
func (env Env) LookupEnv(key string) (value string, ok bool) {
	value, ok = env[key]
	return
}
