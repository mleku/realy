package config

import (
	"os"
	"strings"
)

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

func (env Env) LookupEnv(key string) (value string, ok bool) { value, ok = env[key]; return }
