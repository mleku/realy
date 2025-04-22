// Package timeout provides a simple extension of a net.TCPConn with a
// configurable read/write deadline.
package timeout

import (
	"net"
	"time"

	"realy.mleku.dev/chk"
)

// Conn extends deadline after successful read or write operations
type Conn struct {
	time.Duration
	*net.TCPConn
}

func (c Conn) Read(b []byte) (n int, e error) {
	if n, e = c.TCPConn.Read(b); !chk.E(e) {
		if e = c.SetDeadline(c.getTimeout()); chk.E(e) {
		}
	}
	return
}

func (c Conn) Write(b []byte) (n int, e error) {
	if n, e = c.TCPConn.Write(b); !chk.E(e) {
		if e = c.SetDeadline(c.getTimeout()); chk.E(e) {
		}
	}
	return
}

func (c Conn) getTimeout() (t time.Time) { return time.Now().Add(c.Duration) }
