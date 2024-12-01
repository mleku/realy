package util

import "strings"

func GetKeys(m map[st]st) []st {
	out := make([]st, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func SingleJoiningSlash(a, b st) st {
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
