// Package del is a simple sorted list for database keys, primarily used to
// collect lists of events that need to be deleted either by expiration or for
// the garbage collector.
package del

import "bytes"

type Items [][]byte

func (c Items) Len() int           { return len(c) }
func (c Items) Less(i, j int) bool { return bytes.Compare(c[i], c[j]) < 0 }
func (c Items) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
