package realy

import (
	"sync"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	. "nostr.mleku.dev"
)

type WebSocket struct {
	conn  *websocket.Conn
	mutex sync.Mutex

	// nip42
	challenge S
	authed    S
	limiter   *rate.Limiter
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
