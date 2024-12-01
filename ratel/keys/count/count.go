package count

import (
	"realy.lol/timestamp"
)

type Item struct {
	Serial    uint64
	Size      uint32
	Freshness *timestamp.T
}

type Items []*Item

func (c Items) Len() no         { return len(c) }
func (c Items) Less(i, j no) bo { return c[i].Freshness.I64() < c[j].Freshness.I64() }
func (c Items) Swap(i, j no)    { c[i], c[j] = c[j], c[i] }
func (c Items) Total() (total no) {
	for i := range c {
		total += no(c[i].Size)
	}
	return
}

type ItemsBySerial []*Item

func (c ItemsBySerial) Len() no         { return len(c) }
func (c ItemsBySerial) Less(i, j no) bo { return c[i].Serial < c[j].Serial }
func (c ItemsBySerial) Swap(i, j no)    { c[i], c[j] = c[j], c[i] }
func (c ItemsBySerial) Total() (total no) {
	for i := range c {
		total += no(c[i].Size)
	}
	return
}

type Fresh struct {
	Serial    uint64
	Freshness *timestamp.T
}
type Freshes []*Fresh

func (c Freshes) Len() no         { return len(c) }
func (c Freshes) Less(i, j no) bo { return c[i].Freshness.I64() < c[j].Freshness.I64() }
func (c Freshes) Swap(i, j no)    { c[i], c[j] = c[j], c[i] }
