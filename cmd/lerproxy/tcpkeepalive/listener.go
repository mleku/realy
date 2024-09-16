package tcpkeepalive

import (
	"net"
	"time"

	"mleku.dev/cmd/lerproxy/timeout"
)

// Period can be changed prior to opening a Listener to alter its'
// KeepAlivePeriod.
var Period = 3 * time.Minute

// Listener sets TCP keep-alive timeouts on accepted connections.
// It's used by ListenAndServe and ListenAndServeTLS so dead TCP connections
// (e.g. closing laptop mid-download) eventually go away.
type Listener struct {
	time.Duration
	*net.TCPListener
}

func (ln Listener) Accept() (conn net.Conn, e error) {
	var tc *net.TCPConn
	if tc, e = ln.AcceptTCP(); chk.E(e) {
		return
	}
	if e = tc.SetKeepAlive(true); chk.E(e) {
		return
	}
	if e = tc.SetKeepAlivePeriod(Period); chk.E(e) {
		return
	}
	if ln.Duration != 0 {
		return timeout.Conn{Duration: ln.Duration, TCPConn: tc}, nil
	}
	return tc, nil
}
