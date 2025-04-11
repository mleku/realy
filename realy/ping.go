package realy

import (
	"time"

	"github.com/fasthttp/websocket"

	"realy.mleku.dev/context"
	"realy.mleku.dev/ws"
)

func (s *Server) pinger(ctx context.T, ws *ws.Listener, conn *websocket.Conn,
	ticker *time.Ticker, cancel context.F) {
	defer func() {
		cancel()
		ticker.Stop()
		_ = conn.Close()
	}()
	var err error
	for {
		select {
		case <-ticker.C:
			err = conn.WriteControl(websocket.PingMessage, nil,
				time.Now().Add(s.listeners.WriteWait))
			if err != nil {
				log.E.F("error writing ping: %v; closing websocket", err)
				return
			}
			ws.RealRemote()
		case <-ctx.Done():
			return
		}
	}
}
