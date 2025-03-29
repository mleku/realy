// Package timeout provides a simple extension of a net.TCPConn with a
// configurable read/write deadline.
package timeout

import (
	"net"
	"time"
)

// Conn extends deadline after successful read or write operations
type Conn struct {
	time.Duration
	*net.TCPConn
}

func (c Conn) Read(b by) (n no, e er) {
	if n, e = c.TCPConn.Read(b); !chk.E(e) {
		if e = c.SetDeadline(c.getTimeout()); chk.E(e) {
		}
	}
	return
}

func (c Conn) Write(b by) (n no, e er) {
	if n, e = c.TCPConn.Write(b); !chk.E(e) {
		if e = c.SetDeadline(c.getTimeout()); chk.E(e) {
		}
	}
	return
}

func (c Conn) getTimeout() (t time.Time) { return time.Now().Add(c.Duration) }
