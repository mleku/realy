package socketapi

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"

	"realy.mleku.dev/context"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/realy/interfaces"
	"realy.mleku.dev/ws"
)

type A struct {
	*ws.Listener
	interfaces.Server
	ClientsMu *sync.Mutex
	Clients   map[*websocket.Conn]struct{}
}

func (a *A) Serve(w http.ResponseWriter, r *http.Request, s interfaces.Server) {

	var err error
	ticker := time.NewTicker(s.Listeners().WsPingPeriod)
	ctx, cancel := context.Cancel(s.Context())
	var conn *websocket.Conn
	conn, err = Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	a.ClientsMu.Lock()
	a.Clients[conn] = struct{}{}
	a.ClientsMu.Unlock()
	a.Listener = s.Listeners().GetChallenge(conn, r)

	defer func() {
		cancel()
		ticker.Stop()
		a.ClientsMu.Lock()
		if _, ok := a.Clients[a.Listener.Conn]; ok {
			chk.E(a.Listener.Conn.Close())
			delete(a.Clients, a.Listener.Conn)
			a.Listeners().RemoveSubscriber(a.Listener)
		}
		a.ClientsMu.Unlock()
	}()
	conn.SetReadLimit(a.Listeners().WsMaxMessageSize)
	chk.E(conn.SetReadDeadline(time.Now().Add(a.Listeners().WsPongWait)))
	conn.SetPongHandler(func(string) error {
		chk.E(conn.SetReadDeadline(time.Now().Add(a.Listeners().WsPongWait)))
		return nil
	})
	if a.Server.AuthRequired() {
		a.Listener.RequestAuth()
	}
	if a.Listener.AuthRequested() && len(a.Listener.Authed()) == 0 {
		log.I.F("requesting auth from client from %s", a.Listener.RealRemote())
		if err = authenvelope.NewChallengeWith(a.Listener.Challenge()).Write(a.Listener); chk.E(err) {
			return
		}
		// return
	}
	go a.Pinger(ctx, ticker, cancel, a.Server)
	var message []byte
	var typ int
	for {
		select {
		case <-ctx.Done():
			a.Listener.Close()
			return
		case <-s.Context().Done():
			a.Listener.Close()
			return
		default:
		}
		typ, message, err = conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				log.W.F("unexpected close error from %s: %v",
					a.Listener.Request.Header.Get("X-Forwarded-For"), err)
			}
			return
		}
		if typ == websocket.PingMessage {
			if err = a.Listener.WriteMessage(websocket.PongMessage, nil); chk.E(err) {
			}
			continue
		}
		go a.HandleMessage(message)
	}
}
