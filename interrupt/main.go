package interrupt

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"mleku.dev/atomic"
	"mleku.dev/qu"
)

type HandlerWithSource struct {
	Source string
	Fn     func()
}

var (
	RestartRequested bool // = true
	requested        atomic.Bool

	// ch is used to receive SIGINT (Ctrl+C) signals.
	ch chan os.Signal

	// signals is the list of signals that cause the interrupt
	signals = []os.Signal{os.Interrupt}

	// ShutdownRequestChan is a channel that can receive shutdown requests
	ShutdownRequestChan = qu.T()

	// addHandlerChan is used to add an interrupt handler to the list of handlers to be invoked
	// on SIGINT (Ctrl+C) signals.
	addHandlerChan = make(chan HandlerWithSource)

	// HandlersDone is closed after all interrupt handlers run the first time an interrupt is
	// signaled.
	HandlersDone = make(qu.C)

	interruptCallbacks       []func()
	interruptCallbackSources []string
)

// Listener listens for interrupt signals, registers interrupt callbacks, and responds to custom
// shutdown signals as required
func Listener() {
	invokeCallbacks := func() {
		// run handlers in LIFO order.
		for i := range interruptCallbacks {
			idx := len(interruptCallbacks) - 1 - i
			log.T.F("running callback %d from %s", idx, interruptCallbackSources[idx])
			interruptCallbacks[idx]()
		}
		log.D.Ln("interrupt handlers finished")
		HandlersDone.Q()
		if RestartRequested {
			Restart()
		}
	}
out:
	for {
		select {
		case _ = <-ch:
			fmt.Fprintf(os.Stderr, "\r")
			requested.Store(true)
			invokeCallbacks()
			break out

		case <-ShutdownRequestChan.Wait():
			log.W.Ln("received shutdown request - shutting down...")
			requested.Store(true)
			invokeCallbacks()
			break out

		case handler := <-addHandlerChan:
			interruptCallbacks = append(interruptCallbacks, handler.Fn)
			interruptCallbackSources = append(interruptCallbackSources,
				handler.Source)

		case <-HandlersDone.Wait():
			break out
		}
	}
}

// AddHandler adds a handler to call when a SIGINT (Ctrl+C) is received.
func AddHandler(handler func()) {
	// Create the channel and start the main interrupt handler which invokes all other callbacks
	// and exits if not already done.
	_, loc, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf("%s:%d", loc, line)
	if ch == nil {
		ch = make(chan os.Signal)
		signal.Notify(ch, signals...)
		go Listener()
	}
	addHandlerChan <- HandlerWithSource{
		msg, handler,
	}
}

// Request programmatically requests a shutdown
func Request() {
	_, f, l, _ := runtime.Caller(1)
	log.D.Ln("interrupt requested", f, l, requested.Load())
	if requested.Load() {
		log.D.Ln("requested again")
		return
	}
	requested.Store(true)
	ShutdownRequestChan.Q()
	var ok bool
	select {
	case _, ok = <-ShutdownRequestChan:
	default:
	}
	if ok {
		close(ShutdownRequestChan)
	}
}

// GoroutineDump returns a string with the current goroutine dump in order to show what's going
// on in case of timeout.
func GoroutineDump() string {
	buf := make([]byte, 1<<18)
	n := runtime.Stack(buf, true)
	return string(buf[:n])
}

// RequestRestart sets the reset flag and requests a restart
func RequestRestart() {
	RestartRequested = true
	log.D.Ln("requesting restart")
	Request()
}

// Requested returns true if an interrupt has been requested
func Requested() bool {
	return requested.Load()
}
