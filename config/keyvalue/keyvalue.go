// Package keyvalue provides tools to convert any config struct using go-simpler/env struct
// tagged configuration structures into a sortable slice of key-values, and a printer to render
// them as a bash shell script to set the variables from a configuration file.
package keyvalue

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
)

// KV is a key/value pair.
type KV struct{ Key, Value string }

// KVSlice is a collection of key/value pairs.
type KVSlice []KV

func (kv KVSlice) Len() int           { return len(kv) }
func (kv KVSlice) Less(i, j int) bool { return kv[i].Key < kv[j].Key }
func (kv KVSlice) Swap(i, j int)      { kv[i], kv[j] = kv[j], kv[i] }

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
		var val string
		switch v.(type) {
		case string:
			val = v.(string)
		case int, int64, int32, uint64, uint32, bool, time.Duration:
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

// PrintEnv renders the key/values of a config.C to a provided io.Writer.
func PrintEnv(cfg any, printer io.Writer) {
	_, _ = fmt.Fprintln(printer, "#!/usr/bin/env bash")
	kvs := EnvKV(cfg)
	sort.Sort(kvs)
	for _, v := range kvs {
		_, _ = fmt.Fprintf(printer, "export %s=%s\n", v.Key, v.Value)
	}
}
