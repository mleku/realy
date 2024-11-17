package ratel

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"realy.lol/ratel/keys/count"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/index"
	"realy.lol/ratel/keys/serial"
	"realy.lol/sha256"
	"realy.lol/timestamp"
	"realy.lol/units"
)

const KeyLen = serial.Len + 1
const PrunedLen = sha256.Size + KeyLen
const CounterLen = KeyLen + createdat.Len

// GCCount performs a census of events in the event store. It counts the number
// of events and their size, and if there is a layer 2 enabled, it counts the
// number of events that have been pruned and thus have indexes to count.
//
// Both operations are more efficient combined together rather than separated,
// thus this is a fairly long function.
func (r *T) GCCount() (unpruned, pruned count.Items, unprunedTotal,
	prunedTotal int, err error) {

	//log.D.Ln("running GC count", r.Path())
	overallStart := time.Now()
	prf := []byte{byte(index.Event)}
	evStream := r.DB.NewStream()
	evStream.Prefix = prf
	var countMx sync.Mutex
	var totalCounter int
	evStream.ChooseKey = func(item *badger.Item) (b bool) {
		if item.IsDeletedOrExpired() {
			return
		}
		key := make([]byte, index.Len+serial.Len)
		item.KeyCopy(key)
		ser := serial.FromKey(key)
		size := uint32(item.ValueSize())
		totalCounter++
		countMx.Lock()
		if size == sha256.Size {
			pruned = append(pruned, &count.Item{
				Serial: ser.Uint64(),
				Size:   PrunedLen,
			})
		} else {
			unpruned = append(unpruned, &count.Item{
				Serial: ser.Uint64(),
				Size:   size + KeyLen,
			})
		}
		countMx.Unlock()
		return
	}
	// started := time.Now()
	// run in a background thread to parallelise all the streams
	if err = evStream.Orchestrate(r.Ctx); chk.E(err) {
		return
	}
	log.T.F("counted %d events, %d pruned events in %v %s", len(unpruned),
		len(pruned), time.Now().Sub(overallStart), r.Path())
	var unprunedBySerial, prunedBySerial count.ItemsBySerial
	unprunedBySerial = count.ItemsBySerial(unpruned)
	sort.Sort(unprunedBySerial)
	var countFresh count.Freshes
	// pruneStarted := time.Now()
	counterStream := r.DB.NewStream()
	counterStream.Prefix = []byte{index.Counter.B()}
	v := make([]byte, createdat.Len)
	countFresh = make(count.Freshes, 0, totalCounter)
	counterStream.ChooseKey = func(item *badger.Item) (b bool) {
		key := make([]byte, index.Len+serial.Len)
		item.KeyCopy(key)
		s64 := serial.FromKey(key).Uint64()
		countMx.Lock()
		countFresh = append(countFresh,
			&count.Fresh{
				Serial:    s64,
				Freshness: timestamp.FromUnix(int64(binary.BigEndian.Uint64(v))),
			})
		countMx.Unlock()
		return
	}
	// run in a background thread to parallelise all the streams
	if err = counterStream.Orchestrate(r.Ctx); chk.E(err) {
		return
	}
	// wait until all the jobs are complete
	sort.Sort(countFresh)
	if r.HasL2 {
		// if there is L2 we are marking pruned indexes as well
		// log.I.F("counted %d pruned events in %v %s", len(pruned),
		// 	time.Now().Sub(pruneStarted), r.Path())
		prunedBySerial = count.ItemsBySerial(pruned)
		sort.Sort(prunedBySerial)
	}
	// both slices are now sorted by serial, so we can now iterate the freshness
	// slice and write in the access timestamps to the unpruned
	//
	// this provides the least amount of iteration and computation to essentially
	// zip two tables together
	var unprunedCursor, prunedCursor int
	// we also need to create a map of serials to their respective array index, and
	// we know how big it has to be so we can avoid allocations during the iteration.
	//
	// if there is no L2 this will be an empty map and have nothing added to it.
	prunedMap := make(map[uint64]int, len(prunedBySerial))
	for i := range countFresh {
		// populate freshness of unpruned item
		if len(unprunedBySerial) > i && countFresh[i].Serial ==
			unprunedBySerial[unprunedCursor].Serial {
			// add the counter record to the size
			unprunedBySerial[unprunedCursor].Size += CounterLen
			unprunedBySerial[unprunedCursor].Freshness = countFresh[i].Freshness
			unprunedCursor++
			// if there is no L2 we should not see any here anyway
		} else if r.HasL2 && len(prunedBySerial) > 0 && len(prunedBySerial) < prunedCursor {
			if countFresh[i].Serial ==
				prunedBySerial[prunedCursor].Serial {
				// populate freshness of pruned item
				ps := prunedBySerial[prunedCursor]
				// add the counter record to the size
				ps.Size += CounterLen
				ps.Freshness = countFresh[i].Freshness
				prunedMap[ps.Serial] = prunedCursor
				prunedCursor++
			}
		}
	}
	if r.HasL2 {
		// lastly, we need to count the size of all relevant transactions from the
		// pruned set
		for _, fp := range index.FilterPrefixes {
			// this can all be done concurrently
			go func(fp []byte) {
				evStream = r.DB.NewStream()
				evStream.Prefix = fp
				evStream.ChooseKey = func(item *badger.Item) (b bool) {
					k := item.KeyCopy(nil)
					ser := serial.FromKey(k)
					uSer := ser.Uint64()
					countMx.Lock()
					// the pruned map allows us to (more) directly find the slice index relevant to
					// the serial
					pruned[prunedMap[uSer]].Size += uint32(len(k)) + uint32(item.ValueSize())
					countMx.Unlock()
					return
				}
			}(fp)
		}
	}
	hw, _ := r.GetEventHeadroom()
	unprunedTotal = unpruned.Total()
	up := float64(unprunedTotal)
	var o S
	o += fmt.Sprintf("%8d complete,"+
		"total %0.6f Gb,"+
		"HW %0.6f Gb",
		len(unpruned),
		up/units.Gb,
		float64(hw)/units.Gb,
	)
	if r.HasL2 {
		l2hw, _ := r.GetIndexHeadroom()
		prunedTotal = pruned.Total()
		p := float64(prunedTotal)
		if r.HasL2 {
			o += fmt.Sprintf(",%8d pruned,"+
				"total %0.6f Gb,"+
				"pruned HW %0.6f Gb,computed in %v,%s",
				len(pruned),
				p/units.Gb,
				float64(l2hw)/units.Gb,
				time.Now().Sub(overallStart),
				r.Path(),
			)
		}
	}
	log.D.Ln(o)
	return
}

func (r *T) GetIndexHeadroom() (hw, lw int) {
	limit := r.DBSizeLimit - r.DBSizeLimit*r.DBHighWater/100
	return limit * r.DBHighWater / 100,
		limit * r.DBLowWater / 100
}

func (r *T) GetEventHeadroom() (hw, lw int) {
	return r.DBSizeLimit * r.DBHighWater / 100,
		r.DBSizeLimit * r.DBLowWater / 100
}
