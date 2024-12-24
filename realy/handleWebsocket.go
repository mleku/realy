package realy

import (
	"fmt"
	"net/http"
	"time"

	ws "github.com/fasthttp/websocket"
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
	ip := conn.RemoteAddr().String()
	var realIP st
	if realIP = r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP
	} else if realIP = r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.T.F("connected from %s", ip)
	sock := s.GetChallenge(conn, r, ip)
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.clients[sock.Conn] = struct{}{}
	ticker := time.NewTicker(s.PingPeriod)

	if s.PerConnectionLimiter != nil {
		sock.SetLimiter(rate.NewLimiter(s.PerConnectionLimiter.Limit(),
			s.PerConnectionLimiter.Burst()))
	}
	ctx, cancel := context.Cancel(context.Bg())
	ht := &handleWs{ctx, cancel, ticker, sock, r}
	go s.readLoop(ht)
	go s.ping(ht)
}

type handleWs struct {
	cx
	context.F
	*time.Ticker
	*web.Socket
	*http.Request
}

func (s *Server) readLoop(h *handleWs) {
	var err er
	defer func() {
		h.F()
		h.Stop()
		s.Mutex.Lock()
		if _, ok := s.clients[h.Conn]; ok {
			chk.E(h.Conn.Close())
			delete(s.clients, h.Conn)
			s.RemoveListener(h.Socket)
		}
		s.Mutex.Unlock()
	}()
	h.SetReadLimit(s.MaxMessageSize)
	chk.E(h.SetReadDeadline(time.Now().Add(s.PongWait)))
	h.SetPongHandler(func(st) er {
		chk.E(h.SetReadDeadline(time.Now().Add(s.PongWait)))
		return nil
	})
	if h.AuthRequested() && len(h.Authed()) == 0 {
		log.I.F("requesting auth from client from %s", h.RealRemote())
		if err = authenvelope.NewChallengeWith(h.Challenge()).Write(h.Socket); chk.E(err) {
			return
		}
		return
	}
	var msg by
	var typ no
	for {
		typ, msg, err = h.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err,
				ws.CloseNormalClosure,
				ws.CloseGoingAway,
				ws.CloseNoStatusReceived,
				ws.CloseAbnormalClosure) {
				log.W.F("unexpected close error from %s: %v",
					h.Header.Get("X-Forwarded-For"), err)
			}
			break
		}
		log.T.F("received message\n%s", msg)
		if h.Limiter() != nil {
			if err = h.Limiter().Wait(context.TODO()); chk.T(err) {
				log.W.F("unexpected limiter error %v", err)
				continue
			}
		}
		if typ == ws.PingMessage {
			if err = h.WriteMessage(ws.PongMessage,
				nil); chk.E(err) {
			}
			continue
		}
		go s.handleMessage(h.cx, h.Socket, msg, s.Storage())
	}
}

func (s *Server) ping(h *handleWs) {
	defer func() {
		h.F()
		h.Stop()
		chk.E(h.Conn.Close())
	}()
	var err er
	for {
		select {
		case <-h.C:
			err = h.Conn.WriteControl(ws.PingMessage, nil,
				time.Now().Add(s.WriteWait))
			if err != nil {
				log.E.F("error writing ping: %v; closing websocket", err)
				return
			}
			h.RealRemote()
		case <-h.Done():
			return
		}
	}
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
		if cwh, ok := s.I.(relay.WebSocketHandler); ok {
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
