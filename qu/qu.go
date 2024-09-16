package qu

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"mleku.dev/atomic"
	"mleku.dev/lol"
)

// C is your basic empty struct signalling channel
type C chan struct{}

var (
	createdList                []string
	createdChannels            []C
	createdChannelBufferCounts []int
	mx                         sync.Mutex
	logEnabled                 = atomic.NewBool(false)
)

// SetLogging switches on and off the channel logging
func SetLogging(on bool) {
	logEnabled.Store(on)
}

func l(a ...interface{}) {
	if logEnabled.Load() {
		log.D.Ln(a...)
	}
}

func lc(cl func() string) {
	if logEnabled.Load() {
		log.D.Ln(cl())
	}
}

// T creates an unbuffered chan struct{} for trigger and quit signalling (momentary and breaker
// switches)
func T() C {
	mx.Lock()
	defer mx.Unlock()
	msg := fmt.Sprintf("chan from %s", lol.GetLoc(1))
	l("created", msg)
	createdList = append(createdList, msg)
	o := make(C)
	createdChannels = append(createdChannels, o)
	createdChannelBufferCounts = append(createdChannelBufferCounts, 0)
	return o
}

// Ts creates a buffered chan struct{} which is specifically intended for signalling without
// blocking, generally one is the size of buffer to be used, though there might be conceivable
// cases where the channel should accept more signals without blocking the caller
func Ts(n int) C {
	mx.Lock()
	defer mx.Unlock()
	msg := fmt.Sprintf("buffered chan (%d) from %s", n, lol.GetLoc(1))
	l("created", msg)
	createdList = append(createdList, msg)
	o := make(C, n)
	createdChannels = append(createdChannels, o)
	createdChannelBufferCounts = append(createdChannelBufferCounts, n)
	return o
}

// Q closes the channel, which makes it emit a nil every time it is selected.
func (c C) Q() {
	open := !testChanIsClosed(c)
	lc(func() (o string) {
		lo := getLocForChan(c)
		mx.Lock()
		defer mx.Unlock()
		if open {
			return "closing chan from " + lo + "\n" + strings.Repeat(" ",
				48) + "from" + lol.GetLoc(1)
		} else {
			return "from" + lol.GetLoc(1) + "\n" + strings.Repeat(" ", 48) +
				"channel " + lo + " was already closed"
		}
	},
	)
	if open {
		close(c)
	}
}

// Signal sends struct{}{} on the channel which functions as a momentary switch,
// useful in pairs for stop/start
func (c C) Signal() {
	lc(func() (o string) { return "signalling " + getLocForChan(c) })
	if !testChanIsClosed(c) {
		c <- struct{}{}
	}
}

// Wait should be placed with a `<-` in a select case in addition to the channel
// variable name
func (c C) Wait() <-chan struct{} {
	lc(func() (o string) {
		return fmt.Sprint("waiting on "+getLocForChan(c)+"at",
			lol.GetLoc(1))
	})
	return c
}

// IsClosed exposes a test to see if the channel is closed
func (c C) IsClosed() bool {
	return testChanIsClosed(c)
}

// testChanIsClosed allows you to see whether the channel has been closed so you
// can avoid a panic by trying to close or signal on it
func testChanIsClosed(ch C) (o bool) {
	if ch == nil {
		return true
	}
	select {
	case <-ch:
		o = true
	default:
	}
	return
}

// getLocForChan finds which record connects to the channel in question
func getLocForChan(c C) (s string) {
	s = "not found"
	mx.Lock()
	for i := range createdList {
		if i >= len(createdChannels) {
			break
		}
		if createdChannels[i] == c {
			s = createdList[i]
		}
	}
	mx.Unlock()
	return
}

// once a minute clean up the channel cache to remove closed channels no longer
// in use
func init() {
	go func() {
		for {
			<-time.After(time.Minute)
			l("cleaning up closed channels")
			var c []C
			var ll []string
			mx.Lock()
			for i := range createdChannels {
				if i >= len(createdList) {
					break
				}
				if testChanIsClosed(createdChannels[i]) {
				} else {
					c = append(c, createdChannels[i])
					ll = append(ll, createdList[i])
				}
			}
			createdChannels = c
			createdList = ll
			mx.Unlock()
		}
	}()
}

// PrintChanState creates an output showing the current state of the channels
// being monitored This is a function for use by the programmer while debugging
func PrintChanState() {
	mx.Lock()
	for i := range createdChannels {
		if i >= len(createdList) {
			break
		}
		if testChanIsClosed(createdChannels[i]) {
			log.T.Ln(">>> closed", createdList[i])
		} else {
			log.T.Ln("<<< open", createdList[i])
		}
	}
	mx.Unlock()
}

// GetOpenUnbufferedChanCount returns the number of qu channels that are still open
func GetOpenUnbufferedChanCount() (o int) {
	mx.Lock()
	var c int
	for i := range createdChannels {
		if i >= len(createdChannels) {
			break
		}
		// skip buffered channels
		if createdChannelBufferCounts[i] > 0 {
			continue
		}
		if testChanIsClosed(createdChannels[i]) {
			c++
		} else {
			o++
		}
	}
	mx.Unlock()
	return
}

// GetOpenBufferedChanCount returns the number of qu channels that are still open
func GetOpenBufferedChanCount() (o int) {
	mx.Lock()
	var c int
	for i := range createdChannels {
		if i >= len(createdChannels) {
			break
		}
		// skip unbuffered channels
		if createdChannelBufferCounts[i] < 1 {
			continue
		}
		if testChanIsClosed(createdChannels[i]) {
			c++
		} else {
			o++
		}
	}
	mx.Unlock()
	return
}
