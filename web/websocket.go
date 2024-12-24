package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"

	"realy.lol/atomic"
)

type Socket struct {
	mutex sync.Mutex
	*websocket.Conn
	req           *http.Request
	challenge     atomic.String
	remote        atomic.String
	authed        atomic.String
	authRequested atomic.Bool
	limiter       *rate.Limiter
}

func NewSocket(
	conn *websocket.Conn,
	req *http.Request,
	challenge by,
) (ws *Socket) {
	ws = &Socket{Conn: conn, req: req}
	ws.challenge.Store(st(challenge))
	ws.authRequested.Store(false)
	ws.setRemoteFromReq(req)
	return
}

func (ws *Socket) AuthRequested() bo { return ws.authRequested.Load() }
func (ws *Socket) RequestAuth()      { ws.authRequested.Store(true) }

func (ws *Socket) setRemoteFromReq(r *http.Request) {
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
	ws.remote.Store(rr)
}

func (ws *Socket) Write(p by) (n no, err er) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	err = ws.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		n = len(p)
	}
	return
}

func (ws *Socket) WriteJSON(any interface{}) er {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteJSON(any)
}

func (ws *Socket) WriteMessage(t no, b by) er {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.Conn.WriteMessage(t, b)
}

func (ws *Socket) Challenge() st   { return ws.challenge.Load() }
func (ws *Socket) RealRemote() st  { return ws.remote.Load() }
func (ws *Socket) Authed() st      { return ws.authed.Load() }
func (ws *Socket) AuthedBytes() by { return by(ws.authed.Load()) }
func (ws *Socket) IsAuthed() bo    { return ws.authed.Load() != "" }
func (ws *Socket) SetAuthed(s st) {
	log.T.F("setting authed %0x", s)
	ws.authed.Store(s)
}
func (ws *Socket) Req() *http.Request         { return ws.req }
func (ws *Socket) Limiter() *rate.Limiter     { return ws.limiter }
func (ws *Socket) SetLimiter(l *rate.Limiter) { ws.limiter = l }
