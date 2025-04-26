// Package ws implements nostr websockets with their authentication state.
package ws

import (
	"net/http"
	"strings"
	"sync"

	"github.com/fasthttp/websocket"

	"realy.lol/atomic"
	"realy.lol/log"
)

// Listener is a websocket implementation for a relay listener.
type Listener struct {
	mutex         sync.Mutex
	Conn          *websocket.Conn
	Request       *http.Request
	challenge     atomic.String
	remote        atomic.String
	authed        atomic.String
	authRequested atomic.Bool
}

// NewListener creates a new Listener for listening for inbound connections for a relay.
func NewListener(
	conn *websocket.Conn,
	req *http.Request,
	challenge []byte,
) (ws *Listener) {
	ws = &Listener{Conn: conn, Request: req}
	ws.challenge.Store(string(challenge))
	ws.authRequested.Store(false)
	ws.setRemoteFromReq(req)
	return
}

// AuthRequested returns whether the Listener has asked for auth from the client.
func (ws *Listener) AuthRequested() bool { return ws.authRequested.Load() }

// RequestAuth stores when auth has been required from a client.
func (ws *Listener) RequestAuth() { ws.authRequested.Store(true) }

func (ws *Listener) setRemoteFromReq(r *http.Request) {
	var rr string
	// reverse proxy should populate this field so we see the remote not the proxy
	rem := r.Header.Get("X-Forwarded-For")
	if rem == "" {
		rr = r.RemoteAddr
	} else {
		splitted := strings.Split(rem, " ")
		if len(splitted) == 1 {
			rr = splitted[0]
		}
		if len(splitted) == 2 {
			rr = splitted[1]
		}
		// in case upstream doesn't set this or we are directly listening instead of
		// via reverse proxy or just if the header field is missing, put the
		// connection remote address into the websocket state data.
	}
	if rr == "" {
		// if that fails, fall back to the remote (probably the proxy, unless the realy is
		// actually directly listening)
		rr = ws.Conn.NetConn().RemoteAddr().String()
	}
	ws.remote.Store(rr)
}

// Write a message to send to a client.
func (ws *Listener) Write(p []byte) (n int, err error) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	err = ws.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		n = len(p)
		if strings.Contains(err.Error(), "close sent") {
			// log.I.ToSliceOfBytes("%s", err.Error())
			ws.Close()
			err = nil
			return
		}
	}
	return
}

// WriteJSON encodes whatever into JSON and sends it to the client.
func (ws *Listener) WriteJSON(any interface{}) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteJSON(any)
}

// WriteMessage is a wrapper around the websocket WriteMessage, which includes a websocket
// message type identifier.
func (ws *Listener) WriteMessage(t int, b []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteMessage(t, b)
}

// Challenge returns the current auth challenge string on the socket.
func (ws *Listener) Challenge() string { return ws.challenge.Load() }

// RealRemote returns the stored remote address of the client.
func (ws *Listener) RealRemote() string { return ws.remote.Load() }

// Authed returns the public key the client has authed to the Listener.
func (ws *Listener) Authed() string { return ws.authed.Load() }

// AuthedBytes returns the authed public key that the client has authed to the listener, as a
// byte slice.
func (ws *Listener) AuthedBytes() []byte { return []byte(ws.authed.Load()) }

// IsAuthed returns whether the client has authed to the Listener.
func (ws *Listener) IsAuthed() bool { return ws.authed.Load() != "" }

// SetAuthed loads the pubkey (as a string of the binary pubkey).
func (ws *Listener) SetAuthed(s string) {
	log.T.F("setting authed %0x", s)
	ws.authed.Store(s)
}

// Req returns the http.Request associated with the client connection to the Listener.
func (ws *Listener) Req() *http.Request { return ws.Request }

// Close the Listener connection from the Listener side.
func (ws *Listener) Close() (err error) { return ws.Conn.Close() }
