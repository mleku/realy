package app

import (
	"os"
	"runtime"
	"time"

	"realy.lol/context"
)

func MonitorResources(c context.T) {
	tick := time.NewTicker(time.Minute * 15)
	log.I.Ln("running process", os.Args[0], os.Getpid())
	// memStats := &runtime.MemStats{}
	for {
		select {
		case <-c.Done():
			log.D.Ln("shutting down resource monitor")
			return
		case <-tick.C:
			// runtime.ReadMemStats(memStats)
			log.D.Ln("# goroutines", runtime.NumGoroutine(), "# cgo calls", runtime.NumCgoCall())
			// log.D.S(memStats)
		}
	}
}
