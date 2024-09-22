package web

import (
	"net/http"
	"sync"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"realy.lol/atomic"
)

type Socket struct {
	mutex sync.Mutex
	conn  *websocket.Conn
	req   *http.Request
	// nip42
	challenge atomic.String
	remote    atomic.String
	authed    atomic.String
	limiter   *rate.Limiter
}

func NewSocket(
	conn *websocket.Conn,
	req *http.Request,
	challenge B,
) (ws *Socket) {
	ws = &Socket{conn: conn, req: req}
	ws.challenge.Store(S(challenge))
	return
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

func (ws *Socket) WriteJSON(any interface{}) E {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteJSON(any)
}

func (ws *Socket) WriteMessage(t int, b B) E {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteMessage(t, b)
}

func (ws *Socket) Challenge() S               { return ws.challenge.Load() }
func (ws *Socket) RealRemote() S              { return ws.remote.Load() }
func (ws *Socket) Authed() S                  { return ws.authed.Load() }
func (ws *Socket) SetAuthed(s S)              { ws.authed.Store(s) }
func (ws *Socket) Req() *http.Request         { return ws.req }
func (ws *Socket) Limiter() *rate.Limiter     { return ws.limiter }
func (ws *Socket) SetLimiter(l *rate.Limiter) { ws.limiter = l }
