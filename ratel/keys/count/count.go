// Package count contains a series of data types for managing lists of indexes
// for garbage collection.
package count

import (
	"realy.mleku.dev/timestamp"
)

type Item struct {
	Serial    uint64
	Size      uint32
	Freshness *timestamp.T
}

type Items []*Item

func (c Items) Len() int           { return len(c) }
func (c Items) Less(i, j int) bool { return c[i].Freshness.I64() < c[j].Freshness.I64() }
func (c Items) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Items) Total() (total int) {
	for i := range c {
		total += int(c[i].Size)
	}
	return
}

type ItemsBySerial []*Item

func (c ItemsBySerial) Len() int           { return len(c) }
func (c ItemsBySerial) Less(i, j int) bool { return c[i].Serial < c[j].Serial }
func (c ItemsBySerial) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ItemsBySerial) Total() (total int) {
	for i := range c {
		total += int(c[i].Size)
	}
	return
}

type Fresh struct {
	Serial    uint64
	Freshness *timestamp.T
}
type Freshes []*Fresh

func (c Freshes) Len() int           { return len(c) }
func (c Freshes) Less(i, j int) bool { return c[i].Freshness.I64() < c[j].Freshness.I64() }
func (c Freshes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
