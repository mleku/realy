package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/fasthttp/websocket"

	"realy.lol/atomic"
)

type Socket struct {
	mutex         sync.Mutex
	conn          *websocket.Conn
	req           *http.Request
	challenge     atomic.String
	remote        atomic.String
	authed        atomic.String
	authRequested atomic.Bool
}

func NewSocket(
	conn *websocket.Conn,
	req *http.Request,
	challenge []byte,
) (ws *Socket) {
	ws = &Socket{conn: conn, req: req}
	ws.challenge.Store(string(challenge))
	ws.authRequested.Store(false)
	ws.setRemoteFromReq(req)
	return
}

func (ws *Socket) AuthRequested() bool { return ws.authRequested.Load() }
func (ws *Socket) RequestAuth()        { ws.authRequested.Store(true) }

func (ws *Socket) setRemoteFromReq(r *http.Request) {
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
		rr = ws.conn.NetConn().RemoteAddr().String()
	}
	ws.remote.Store(rr)
}

func (ws *Socket) Write(p []byte) (n int, err error) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	err = ws.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		n = len(p)
	}
	return
}

func (ws *Socket) WriteJSON(any interface{}) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteJSON(any)
}

func (ws *Socket) WriteMessage(t int, b []byte) error {
	ws.mutex.Lock()
	// defer ws.mutex.Unlock()
	return ws.conn.WriteMessage(t, b)
}

func (ws *Socket) Challenge() string   { return ws.challenge.Load() }
func (ws *Socket) RealRemote() string  { return ws.remote.Load() }
func (ws *Socket) Authed() string      { return ws.authed.Load() }
func (ws *Socket) AuthedBytes() []byte { return []byte(ws.authed.Load()) }
func (ws *Socket) IsAuthed() bool      { return ws.authed.Load() != "" }
func (ws *Socket) SetAuthed(s string) {
	log.T.F("setting authed %0x", s)
	ws.authed.Store(s)
}
func (ws *Socket) Req() *http.Request { return ws.req }
