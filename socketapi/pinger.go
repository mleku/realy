package socketapi

import (
	"time"

	"github.com/fasthttp/websocket"

	"realy.lol/context"
	"realy.lol/log"
)

func (a *A) Pinger(ctx context.T, ticker *time.Ticker, cancel context.F, remote string) {
	log.T.F("running pinger for %s", remote)
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
				time.Now().Add(DefaultPingWait))
			if err != nil {
				log.E.F("%s error writing ping: %v; closing websocket", remote, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
