package realy

import (
	"fmt"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"

	"realy.lol/context"
	"realy.lol/envelopes"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closeenvelope"
	"realy.lol/envelopes/countenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/noticeenvelope"
	"realy.lol/envelopes/reqenvelope"
	"realy.lol/realy/api"
	"realy.lol/realy/listeners"
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/web"
)

func (s *Server) handleWebsocket(h api.H) {
	conn, err := listeners.Upgrader.Upgrade(h.ResponseWriter, h.Request, nil)
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
	if realIP = h.Request.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP
	} else if realIP = h.Request.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.T.F("connected from %s", ip)
	ws := s.listeners.GetChallenge(conn, h.Request, ip)
	if s.options.PerConnectionLimiter != nil {
		// this does not apply to users on the owners' Followed list
		ws.SetLimiter(rate.NewLimiter(s.options.PerConnectionLimiter.Limit(),
			s.options.PerConnectionLimiter.Burst()))
	}
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
				s.listeners.RemoveListener(ws)
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
						h.Request.Header.Get("X-Forwarded-For"), err)
				}
				break
			}
			// if ws.Limiter() != nil {
			// 	if err = ws.Limiter().Wait(context.TODO()); chk.T(err) {
			// 		log.W.F("unexpected limiter error %v", err)
			// 		continue
			// 	}
			// }
			if typ == websocket.PingMessage {
				if err = ws.WriteMessage(websocket.PongMessage, nil); chk.E(err) {
				}
				continue
			}
			go s.handleMessage(ctx, ws, message, sto)
		}
	}()
	go s.pinger(ctx, ws, conn, ticker, cancel)
}

func (s *Server) pinger(ctx context.T, ws *web.Socket, conn *websocket.Conn, ticker *time.Ticker, cancel context.F) {
	defer func() {
		cancel()
		ticker.Stop()
		conn.Close()
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

func (s *Server) handleMessage(c context.T, ws *web.Socket, msg []byte, sto store.I) {
	var notice []byte
	var err error
	var t string
	var rem []byte
	if t, rem, err = envelopes.Identify(msg); chk.E(err) {
		notice = []byte(err.Error())
	}
	switch t {
	case eventenvelope.L:
		notice = s.handleEvent(c, ws, rem, sto)
	case countenvelope.L:
		notice = s.handleCount(c, ws, rem, sto)
	case reqenvelope.L:
		notice = s.handleReq(c, ws, rem, sto)
	case closeenvelope.L:
		notice = s.handleClose(ws, rem)
	case authenvelope.L:
		notice = s.handleAuth(ws, rem)
	default:
		if cwh, ok := s.relay.(relay.WebSocketHandler); ok {
			cwh.HandleUnknownType(ws, t, rem)
		} else {
			notice = []byte(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
		}
	}
	if len(notice) > 0 {
		log.D.F("notice->%s %s", ws.RealRemote(), notice)
		if err = noticeenvelope.NewFrom(notice).Write(ws); err != nil {
			return
		}
	}
}
