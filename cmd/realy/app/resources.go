package app

import (
	"os"
	"runtime"
	"time"
)

func MonitorResources(c cx) {
	tick := time.NewTicker(time.Minute)
	log.I.Ln("running process", os.Args[0], os.Getpid())
	// memStats := &runtime.MemStats{}
	for {
		select {
		case <-c.Done():
			log.D.Ln("shutting down resource monitor")
			return
		case <-tick.C:
			// runtime.ReadMemStats(memStats)
			log.D.Ln("# goroutines", runtime.NumGoroutine())
			log.D.Ln("# cgo calls", runtime.NumCgoCall())
			// log.D.S(memStats)
		}
	}
}
