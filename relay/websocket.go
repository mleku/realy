package relay

import (
	"net/http"
	"sync"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"realy.lol/atomic"
)

type WebSocket struct {
	conn  *websocket.Conn
	mutex sync.Mutex
	req   *http.Request
	// nip42
	challenge atomic.String
	Remote    atomic.String
	authed    B
	limiter   *rate.Limiter
}

func (ws *WebSocket) Write(p []byte) (n int, err error) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	err = ws.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		n = len(p)
	}
	return
}

func (ws *WebSocket) WriteJSON(any interface{}) E {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteJSON(any)
}

func (ws *WebSocket) WriteMessage(t int, b B) E {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	return ws.conn.WriteMessage(t, b)
}
