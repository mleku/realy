package main

import (
	"os"
	"sync"
	"time"

	"lukechampine.com/frand"

	"realy.lol/bech32encoding"
	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/interrupt"
	"realy.lol/keys"
	"realy.lol/layer2"
	"realy.lol/lol"
	"realy.lol/qu"
	"realy.lol/ratel"
	"realy.lol/tag"
	"realy.lol/tests"
	"realy.lol/units"
)

type Counter struct {
	id        by
	size      no
	requested no
}

func main() {
	lol.NoTimeStomp.Store(true)
	lol.SetLogLevel(lol.LevelNames[lol.Debug])
	var (
		err            er
		sec            by
		mx             sync.Mutex
		counter        []Counter
		total          no
		MaxContentSize = units.Mb / 2
		TotalSize      = 1
		MaxDelay       = time.Second / 40
		HW             = 50
		LW             = 25
		// fill rate capped to size of difference between high and low water mark
		diff = TotalSize * units.Gb * (HW - LW) / 100
	)
	if sec, err = keys.GenerateSecretKey(); chk.E(err) {
		panic(err)
	}
	var nsec by
	if nsec, err = bech32encoding.HexToNsec(sec); chk.E(err) {
		panic(err)
	}
	log.T.Ln("signing with", nsec)
	c, cancel := context.Cancel(context.Bg())
	var wg sync.WaitGroup
	// defer cancel()
	// create L1 with cache management settings enabled; we do it in the current dir
	// because os.TempDir can point to a ramdisk which is very impractical for this
	// test.
	path := "./badgerbadgertest"
	os.RemoveAll(path)
	b1 := ratel.GetBackend(c, &wg, true, units.Gb, lol.Error, 4*units.Mb,
		TotalSize, LW, HW, 2)
	// create L2 with no cache management
	b2 := ratel.GetBackend(c, &wg, false, units.Gb, lol.Trace, 4*units.Mb)
	// Respond to interrupt signal and clean up after interrupt or end of test.
	// defer chk.E(os.RemoveAll(path))
	interrupt.AddHandler(func() {
		cancel()
		chk.E(os.RemoveAll(path))
	})
	// now join them together in a 2 level eventstore
	twoLevel := layer2.Backend{
		Ctx: c,
		WG:  &wg,
		L1:  b1,
		L2:  b2,
	}
	if err = twoLevel.Init(path); chk.E(err) {
		os.Exit(1)
	}
	// start GC
	// go b1.GarbageCollector()
end:
	for {
		select {
		case <-c.Done():
			log.I.Ln("context canceled")
			return
		default:
		}
		mx.Lock()
		if total > TotalSize*10*units.Gb {
			log.I.Ln(total, TotalSize*10*units.Gb)
			mx.Unlock()
			cancel()
			return
		}
		mx.Unlock()
		newEvent := qu.T()
		go func() {
			ticker := time.NewTicker(time.Second)
			var fetchIDs []by
			// start fetching loop
			for {
				select {
				case <-newEvent:
					// make new request, not necessarily from existing... bias rng
					// factor by request count
					mx.Lock()
					var sum no
					for i := range counter {
						rn := frand.Intn(256)
						if sum > diff {
							// don't overfill
							break
						}
						// multiply this number by the number of accesses the event
						// has and request every event that gets over 50% so that we
						// create a bias towards already requested.
						if counter[i].requested+rn > 216 {
							log.T.Ln("counter", counter[i].requested, "+", rn,
								"=",
								counter[i].requested+rn)
							// log.T.Ln("adding to fetchIDs")
							counter[i].requested++
							fetchIDs = append(fetchIDs, counter[i].id)
							sum += counter[i].size
						}
					}
					// if len(fetchIDs) > 0 {
					//	log.T.Ln("fetchIDs", len(fetchIDs), fetchIDs)
					// }
					mx.Unlock()
				case <-ticker.C:
					// copy out current list of events to request
					mx.Lock()
					log.T.Ln("ticker", len(fetchIDs))
					ids := tag.NewWithCap(len(fetchIDs))
					for i := range fetchIDs {
						ids.Append(fetchIDs[i])
					}
					fetchIDs = fetchIDs[:0]
					mx.Unlock()
					if ids.Len() > 0 {
						_, err = twoLevel.QueryEvents(c, &filter.T{IDs: ids})
					}
				case <-c.Done():
					log.I.Ln("context canceled")
					return
				}
			}
		}()
		var ev *event.T
		var bs no
	out:
		for {
			select {
			case <-c.Done():
				log.I.Ln("context canceled")
				return
			default:
			}
			if ev, bs, err = tests.GenerateEvent(MaxContentSize); chk.E(err) {
				return
			}
			mx.Lock()
			counter = append(counter, Counter{id: ev.ID, size: bs, requested: 1})
			total += bs
			if total > TotalSize*10*units.Gb {
				log.I.Ln(total, TotalSize*units.Gb)
				mx.Unlock()
				cancel()
				break out
			}
			mx.Unlock()
			newEvent.Signal()
			sc, _ := context.Timeout(c, 2*time.Second)
			if err = twoLevel.SaveEvent(sc, ev); chk.E(err) {
				continue end
			}
			delay := frand.Intn(no(MaxDelay))
			log.T.Ln("waiting between", delay, "ns")
			if delay == 0 {
				continue
			}
			select {
			case <-c.Done():
				log.I.Ln("context canceled")
				return
			case <-time.After(time.Duration(delay)):
			}
		}
		select {
		case <-c.Done():
		}
	}
}
