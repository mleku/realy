package socketapi

import (
	"net/http"

	"github.com/fasthttp/websocket"
)

var Upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}
