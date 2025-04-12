package socketapi

import (
	"time"

	"github.com/fasthttp/websocket"

	"realy.mleku.dev/context"
	"realy.mleku.dev/realy/interfaces"
)

func (a *A) Pinger(ctx context.T, ticker *time.Ticker, cancel context.F, s interfaces.Server) {
	defer func() {
		cancel()
		ticker.Stop()
		_ = a.Listener.Conn.Close()
	}()
	var err error
	for {
		select {
		case <-ticker.C:
			err = a.Listener.Conn.WriteControl(websocket.PingMessage, nil,
				time.Now().Add(s.Listeners().WsPingWait))
			if err != nil {
				log.E.F("error writing ping: %v; closing websocket", err)
				return
			}
			a.Listener.RealRemote()
		case <-ctx.Done():
			return
		}
	}
}
