package realy

import (
	"fmt"
	"net/http"
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
	"realy.lol/realy/listeners"
	"realy.lol/relay"
	"realy.lol/store"
	"realy.lol/web"
)

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := listeners.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[conn] = struct{}{}
	ticker := time.NewTicker(s.listeners.PingPeriod)
	ip := conn.RemoteAddr().String()
	var realIP st
	if realIP = r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP
	} else if realIP = r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.T.F("connected from %s", ip)
	ws := s.listeners.GetChallenge(conn, r, ip)
	if s.options.PerConnectionLimiter != nil {
		ws.SetLimiter(rate.NewLimiter(s.options.PerConnectionLimiter.Limit(),
			s.options.PerConnectionLimiter.Burst()))
	}
	ctx, cancel := context.Cancel(context.Bg())
	sto := s.relay.Storage(ctx)
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
		conn.SetPongHandler(func(st) er {
			chk.E(conn.SetReadDeadline(time.Now().Add(s.listeners.PongWait)))
			return nil
		})
		if ws.AuthRequested() && len(ws.Authed()) == 0 {
			log.I.F("requesting auth from client from %s", ws.RealRemote())
			if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
				return
			}
			return
		}
		var message by
		var typ no
		for {
			typ, message, err = conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure,
					websocket.CloseGoingAway, websocket.CloseNoStatusReceived,
					websocket.CloseAbnormalClosure) {
					log.W.F("unexpected close error from %s: %v",
						r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}
			if ws.Limiter() != nil {
				if err := ws.Limiter().Wait(context.TODO()); chk.T(err) {
					log.W.F("unexpected limiter error %v", err)
					continue
				}
			}
			if typ == websocket.PingMessage {
				if err = ws.WriteMessage(websocket.PongMessage, nil); chk.E(err) {
				}
				continue
			}
			go s.handleMessage(ctx, ws, message, sto)
		}
	}()
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			chk.E(conn.Close())
		}()
		var err er
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
	}()
}

func (s *Server) handleMessage(c cx, ws *web.Socket, msg by, sto store.I) {
	var notice by
	var err er
	var t st
	var rem by
	if t, rem, err = envelopes.Identify(msg); chk.E(err) {
		notice = by(err.Error())
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
			notice = by(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
		}
	}
	if len(notice) > 0 {
		log.D.F("notice %s", notice)
		if err = noticeenvelope.NewFrom(notice).Write(ws); chk.E(err) {
		}
	}
}
