package ratel

import (
	"sort"

	"realy.lol/ratel/keys/count"
	"realy.lol/units"
)

type DelItems []uint64

// GCMark first gathers the serial, data size and last accessed information
// about all events and pruned events using GCCount then sorts the results of
// the events and indexes by least recently accessed and generates the set of
// serials of events that need to be deleted
func (r *T) GCMark() (pruneEvents, pruneIndexes DelItems, err error) {
	var unpruned, pruned count.Items
	var uTotal, pTotal int
	if unpruned, pruned, uTotal, pTotal, err = r.GCCount(); chk.E(err) {
		return
	}
	hw, lw := r.GetEventHeadroom()
	if uTotal > hw {
		// run event GC mark
		sort.Sort(unpruned)
		pruneOff := uTotal - lw
		var cumulative, lastIndex int
		for lastIndex = range unpruned {
			if cumulative > pruneOff {
				break
			}
			cumulative += int(unpruned[lastIndex].Size)
			pruneEvents = append(pruneEvents, unpruned[lastIndex].Serial)
		}
		log.D.F("found %d events to prune,which will bring current "+
			"utilization down to %0.6f Gb,%s",
			lastIndex-1, float64(uTotal-cumulative)/units.Gb, r.Path())
	}
	l2hw, l2lw := r.GetIndexHeadroom()
	if r.HasL2 && pTotal > l2hw {
		// run index GC mark
		sort.Sort(pruned)
		var lastIndex int
		// we want to remove the oldest indexes until at or below the index low water mark.
		space := pTotal
		// count the number of events until the low water mark
		for lastIndex = range pruned {
			if space < l2lw {
				break
			}
			space -= int(pruned[lastIndex].Size)
		}
		log.D.F("deleting %d indexes using %d bytes to bring pruned index size to %d",
			lastIndex+1, pTotal-l2lw, space)
		for i := range pruned {
			if i > lastIndex {
				break
			}
			pruneIndexes = append(pruneIndexes, pruned[i].Serial)
		}
	}
	return
}
