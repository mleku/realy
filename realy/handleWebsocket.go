package realy

import (
	"net/http"
	"time"

	"github.com/fasthttp/websocket"

	"realy.mleku.dev/context"
	"realy.mleku.dev/envelopes/authenvelope"
	"realy.mleku.dev/realy/subscribers"
	"realy.mleku.dev/socketapi"
)

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := subscribers.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[conn] = struct{}{}
	ticker := time.NewTicker(s.listeners.PingPeriod)
	ip := conn.RemoteAddr().String()
	var realIP string
	if realIP = r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP
	} else if realIP = r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.T.F("connected from %s", ip)
	ws := s.listeners.GetChallenge(conn, r, ip)
	ctx, cancel := context.Cancel(context.Bg())
	sto := s.relay.Storage()
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			s.clientsMu.Lock()
			if _, ok := s.clients[conn]; ok {
				chk.E(conn.Close())
				delete(s.clients, conn)
				s.listeners.RemoveSubscriber(ws)
			}
			s.clientsMu.Unlock()
		}()
		conn.SetReadLimit(s.listeners.MaxMessageSize)
		chk.E(conn.SetReadDeadline(time.Now().Add(s.listeners.PongWait)))
		conn.SetPongHandler(func(string) error {
			chk.E(conn.SetReadDeadline(time.Now().Add(s.listeners.PongWait)))
			return nil
		})
		if s.authRequired {
			ws.RequestAuth()
		}
		if ws.AuthRequested() && len(ws.Authed()) == 0 {
			log.I.F("requesting auth from client from %s", ws.RealRemote())
			if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
				return
			}
			// return
		}
		var message []byte
		var typ int
		for {
			typ, message, err = conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived,
					websocket.CloseAbnormalClosure,
				) {
					log.W.F("unexpected close error from %s: %v",
						r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}
			if typ == websocket.PingMessage {
				if err = ws.WriteMessage(websocket.PongMessage, nil); chk.E(err) {
				}
				continue
			}
			a := &socketapi.A{ws}
			go s.handleMessage(ctx, a, message, sto)
		}
	}()
	go s.pinger(ctx, ws, conn, ticker, cancel)
}
