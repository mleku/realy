package ws

import (
	"crypto/rand"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"realy.lol/atomic"
	"realy.lol/bech32encoding"
	"realy.lol/context"
	"realy.lol/ec/bech32"
	"realy.lol/qu"
)

type MessageType int

// Serv is a wrapper around a fasthttp/websocket with mutex locking and NIP-42 IsAuthed support
// for handling inbound connections from clients.
type Serv struct {
	Ctx       Ctx
	Cancel    context.F
	Conn      *websocket.Conn
	remote    atomic.String
	mutex     sync.Mutex
	Request   *http.Request // original request
	challenge atomic.String // nip42
	Pending   atomic.Value  // for DM CLI authentication
	authPub   atomic.Value
	Authed    qu.C
}

func New(c Ctx, conn *websocket.Conn, r *http.Request, maxMsg int) (ws *Serv) {
	// authPubKey must be initialized with a zero length slice so it can be detected when it
	// hasn't been loaded.
	var authPubKey atomic.Value
	authPubKey.Store(B{})
	ws = &Serv{Ctx: c, Conn: conn, Request: r, Authed: qu.T(), authPub: authPubKey}
	ws.generateChallenge()
	ws.setRemoteFromReq(r)
	conn.SetReadLimit(int64(maxMsg))
	conn.EnableWriteCompression(true)
	return
}

// Ping sends a ping to see if the other side is still responsive.
func (ws *Serv) Ping() (err E) { return ws.write(websocket.PingMessage, nil) }

// Pong sends a Pong message, should be the response to receiving  Ping.
func (ws *Serv) Pong() (err E) { return ws.write(websocket.PongMessage, nil) }

// Close signals the other side to close the connection.
func (ws *Serv) Close() (err E) { return ws.write(websocket.CloseMessage, nil) }

// Challenge returns the current challenge on a websocket.
func (ws *Serv) Challenge() (challenge B) { return B(ws.challenge.Load()) }

// Remote returns the current real remote.
func (ws *Serv) Remote() (remote S) { return ws.remote.Load() }

// setRemote sets the current remote URL that is returned by Remote.
func (ws *Serv) setRemote(remote S) { ws.remote.Store(remote) }

// write writes a message with a given websocket type specifier
func (ws *Serv) write(t MessageType, b B) (err E) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if len(b) != 0 {
		log.D.F("sending message to %s %0x\n%s", ws.Remote(), ws.AuthPub(), string(b))
	}
	chk.E(ws.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5)))
	return ws.Conn.WriteMessage(int(t), b)
}

// WriteTextMessage writes a text (binary?) message
func (ws *Serv) WriteTextMessage(b B) (err E) { return ws.write(websocket.TextMessage, b) }

// Write implements the standard io.Writer interface.
func (ws *Serv) Write(b []byte) (n int, err error) {
	if err = ws.WriteTextMessage(b); chk.E(err) {
		return
	}
	n = len(b)
	return
}

var _ io.Writer = (*Serv)(nil)

const ChallengeLength = 16
const ChallengeHRP = "nchal"

// generateChallenge gathers new entropy to generate a new challenge, stores it and returns it.
func (ws *Serv) generateChallenge() (challenge S) {
	var err error
	// create a new challenge for this connection
	cb := make([]byte, ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		// i never know what to do for this case, panic? usually just ignore, it should never happen
		panic(err)
	}
	var b5 B
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded B
	if encoded, err = bech32.Encode(bech32.B(ChallengeHRP), b5); chk.E(err) {
		return
	}
	challenge = S(encoded)
	ws.challenge.Store(challenge)
	return
}

// setAuthPub loads the authPubKey atomic of the websocket.
func (ws *Serv) setAuthPub(a B) {
	aa := make(B, 0, len(a))
	copy(aa, a)
	ws.authPub.Store(aa)
}

// AuthPub returns the current authed Pubkey.
func (ws *Serv) AuthPub() (a B) {
	b := ws.authPub.Load().(B)
	a = make(B, 0, len(b))
	// make a copy because bytes are references
	a = append(a, b...)
	return
}

func (ws *Serv) HasAuth() bool {
	b := ws.authPub.Load().(B)
	return len(b) > 0
}

func (ws *Serv) setRemoteFromReq(r *http.Request) {
	var rr string
	// reverse proxy should populate this field so we see the remote not the proxy
	rem := r.Header.Get("X-Forwarded-For")
	if rem != "" {
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
		if rr == "" {
			rr = r.RemoteAddr
		}
	} else {
		// if that fails, fall back to the remote (probably the proxy, unless the realy is
		// actually directly listening)
		rr = ws.Conn.NetConn().RemoteAddr().String()
	}
	ws.setRemote(rr)
}
