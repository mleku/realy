// Package util provides some helpers for lerproxy, a tool to convert maps of
// strings to slices of the same strings, and a helper to avoid putting two / in
// a URL.
package util

import "strings"

func GetKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func SingleJoiningSlash(a, b string) string {
	suffixSlash := strings.HasSuffix(a, "/")
	prefixSlash := strings.HasPrefix(b, "/")
	switch {
	case suffixSlash && prefixSlash:
		return a + b[1:]
	case !suffixSlash && !prefixSlash:
		return a + "/" + b
	}
	return a + b
}
