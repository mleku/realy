package realy

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/fasthttp/websocket"
	"golang.org/x/time/rate"
	"realy.lol/auth"
	"realy.lol/bech32encoding"
	"realy.lol/ec/bech32"
	"realy.lol/envelopes"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closedenvelope"
	"realy.lol/envelopes/closeenvelope"
	"realy.lol/envelopes/countenvelope"
	"realy.lol/envelopes/eoseenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/noticeenvelope"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/envelopes/reqenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/ints"
	"realy.lol/kind"
	"realy.lol/normalize"
	"realy.lol/relay"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

// TODO: consider moving these to Server as config params
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait / 2

	// Maximum message size allowed from peer.
	maxMessageSize = 512000
)

// TODO: consider moving these to Server as config params
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const ChallengeLength = 16
const ChallengeHRP = "nchal"

func challenge(conn *websocket.Conn, req *http.Request, addr string) (ws *web.Socket) {
	var err error
	// create a new challenge for this connection
	cb := make([]byte, ChallengeLength)
	if _, err = rand.Read(cb); chk.E(err) {
		// i never know what to do for this case, panic? usually just ignore, it should never happen
		panic(err)
	}
	var b5 B
	if b5, err = bech32encoding.ConvertForBech32(cb); chk.E(err) {
		return
	}
	var encoded B
	if encoded, err = bech32.Encode(bech32.B(ChallengeHRP), b5); chk.E(err) {
		return
	}
	ws = web.NewSocket(conn, req, encoded)
	return
}

func (s *Server) handleMessage(c Ctx, ws *web.Socket, msg B, sto store.I) {
	var notice B
	var err E
	var t S
	var rem B
	if t, rem, err = envelopes.Identify(msg); chk.E(err) {
		notice = B(err.Error())
	}
	switch t {
	case eventenvelope.L:
		notice = s.doEvent(c, ws, rem, sto)
	case countenvelope.L:
		notice = s.doCount(c, ws, rem, sto)
	case reqenvelope.L:
		notice = s.doReq(c, ws, rem, sto)
	case closeenvelope.L:
		notice = s.doClose(c, ws, rem, sto)
	case authenvelope.L:
		notice = s.doAuth(c, ws, rem, sto)
	default:
		if cwh, ok := s.relay.(relay.WebSocketHandler); ok {
			cwh.HandleUnknownType(ws, t, rem)
		} else {
			notice = B(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
		}
	}
	if len(notice) > 0 {
		log.D.F("notice %s", notice)
		if err = noticeenvelope.NewFrom(notice).Write(ws); chk.E(err) {
		}
	}
}

func (s *Server) doEvent(c Ctx, ws *web.Socket, req B, sto store.I) (msg B) {
	log.D.F("doEvent %s %s", ws.RealRemote(), req)
	var err E
	var ok bool
	var rem B
	advancedDeleter, _ := sto.(relay.AdvancedDeleter)

	env := eventenvelope.NewSubmission()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}

	if !s.relay.AcceptEvent(c, env.T, ws.Req(), B(ws.Authed())) {
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok &&
			auther.AuthEnabled() && !ws.AuthRequested() {

			if err = okenvelope.NewFrom(env.ID, false,
				normalize.AuthRequired.F("auth required for count request processing")).
				Write(ws); chk.E(err) {
			}
			log.D.F("requesting auth from client")
			if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
				return
			}
			return
		}
		log.W.F("rejecting event\n%s", env.T.Serialize())
		return
	}
	// check id
	if !equals(env.GetIDBytes(), env.ID) {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F("event id is computed incorrectly")).Write(ws); chk.E(err) {
			return
		}
		return
	}

	// check signature
	if ok, err = env.Verify(); err != nil {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("failed to verify signature")).Write(ws); chk.E(err) {
			return
		}
	} else if !ok {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("signature is invalid")).Write(ws); chk.E(err) {
			return
		}
		return
	}
	if env.T.Kind.K == kind.Deletion.K {
		log.I.F("delete event\n%s", env.T.Serialize())
		// event deletion -- nip09
		for _, t := range env.Tags.Value() {
			var res []*event.T
			if t.Len() >= 2 {
				switch {
				case equals(t.Key(), B("e")):
					evId := make(B, sha256.Size)
					if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
						continue
					}
					// fetch event to be deleted
					res, err = s.relay.Storage(c).
						QueryEvents(c, &filter.T{IDs: tag.New(evId)})
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
				case equals(t.Key(), B("a")):
					split := bytes.Split(t.Value(), B{':'})
					if len(split) != 3 {
						continue
					}
					kin := ints.New(uint16(0))
					if _, err = kin.UnmarshalJSON(split[0]); chk.E(err) {
						return
					}
					f := filter.New()
					f.Kinds.K = []*kind.T{kind.New(kin.Uint16())}
					aut := make(B, 0, len(split[1])/2)
					if aut, err = hex.DecAppend(aut, split[1]); chk.E(err) {
						return
					}
					f.Authors.Append(aut)
					f.Tags.AppendTags(tag.New(B{'#', 'd'}, split[2]))
					res, err = s.relay.Storage(c).QueryEvents(c, f)
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
				}
			}
			if len(res) < 1 {
				// this will happen if event is not in the database
				continue
			}
			for _, target := range res {
				if target.Kind.K == kind.Deletion.K {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F("cannot delete delete event %s",
							env.ID)).Write(ws); chk.E(err) {
						return
					}
				}
				if target.CreatedAt.Int() > env.T.CreatedAt.Int() {
					log.I.F("not replacing\n%d%\nbecause delete event is older\n%d",
						target.CreatedAt.Int(), env.T.CreatedAt.Int())
					continue
				}
				// check if this can be deleted
				if !equals(target.PubKey, env.PubKey) {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F("only author can delete event")).Write(ws); chk.E(err) {
						return
					}
					return
				}

				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(c, t.Value(), env.PubKey)
				}

				// delete the event
				if err = sto.DeleteEvent(c, target.EventID()); err != nil {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
						return
					}
					return
				}

				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.Value(), env.PubKey)
				}
			}
		}
		if err = okenvelope.NewFrom(env.ID, true).Write(ws); chk.E(err) {
			return
		}
		// if the event is a delete we still want to save it.
	}
	ok, reason := AddEvent(c, s.relay, env.T, ws.Req(), B(ws.Authed()))
	if err = okenvelope.NewFrom(env.ID, ok, reason).Write(ws); chk.E(err) {
		return
	}
	return
}

func (s *Server) doCount(c context.Context, ws *web.Socket, req B,
	store store.I) (msg B) {

	counter, ok := store.(relay.EventCounter)
	if !ok {
		return normalize.Restricted.F("this relay does not support NIP-45")
	}

	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		// ignore requests until request is responded to
		return
	}

	var err E
	var rem B
	env := countenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}

	if env.ID == nil || env.ID.String() == "" {
		return normalize.Error.F("COUNT has no <id>")
	}

	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		if !accepter.AcceptReq(c, ws.Req(), env.ID.T, env.Filters, B(ws.Authed())) {

			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok &&
				auther.AuthEnabled() && !ws.AuthRequested() {

				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.ID,
					normalize.AuthRequired.F("auth required for count processing")).
					Write(ws); chk.E(err) {
				}
				log.I.F("requesting auth from client from %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}
	}

	var total N
	for _, f := range env.Filters.F {
		// prevent kind-4 events from being returned to unauthed users, only when
		// authentication is a thing
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.Contains(kind.EncryptedDirectMessage) {
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("p"))
				switch {
				case len(ws.Authed()) == 0:
					// not authenticated
					return normalize.Restricted.F(
						"this realy does not serve kind-4 to unauthenticated users," +
							" does your client implement NIP-42?")
				case senders.Len() == 1 &&
					receivers.Len() < 2 &&
					equals(senders.F()[0], B(ws.Authed())):
					// allowed filter: ws.authed is sole sender (filter specifies one or all
					// receivers)
				case receivers.Len() == 1 &&
					senders.Len() < 2 &&
					equals(receivers.N(0).Value(), B(ws.Authed())):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all
					// senders)
				default:
					// restricted filter: do not return any events, even if other elements in
					// filters array were not restricted). client should know better.
					return normalize.Restricted.F("authenticated user does not have" +
						" authorization for requested filters")
				}
			}
		}
		var count N
		count, err = counter.CountEvents(c, f)
		if err != nil {
			log.E.F("store: %v", err)
			continue
		}
		total += count
	}
	var res *countenvelope.Response
	if res, err = countenvelope.NewResponseFrom(env.ID.String(), N(total),
		false); chk.E(err) {
		return
	}
	if err = res.Write(ws); chk.E(err) {
		return
	}
	return
}

func (s *Server) doReq(c Ctx, ws *web.Socket, req B, sto store.I) (r B) {

	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		// ignore requests until request is responded to
		return
	}
	var err E
	var rem B
	env := reqenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}

	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		if !accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters, B(ws.Authed())) {

			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok &&
				auther.AuthEnabled() && !ws.AuthRequested() {

				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).
					Write(ws); chk.E(err) {
				}
				log.I.F("requesting auth from client from %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}
	}

	for _, f := range env.Filters.F {
		var i uint
		if filter.Present(f.Limit) {
			if *f.Limit == 0 {
				continue
			}
			i = *f.Limit
		}
		// prevent kind-4 events from being returned to unauthed users,
		//   only when authentication is a thing
		if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.Contains(kind.EncryptedDirectMessage) {
				// if slices.Contains(f.Kinds.K, kind.EncryptedDirectMessage) {
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("p"))
				switch {
				case len(ws.Authed()) == 0:
					ws.RequestAuth()
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for request processing")).
						Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					// not authenticated
					notice := normalize.Restricted.F(
						"this realy does not serve kind-4 to unauthenticated users," +
							" does your client implement NIP-42?")
					return notice
				case senders.Len() == 1 &&
					receivers.Len() < 2 &&
					equals(senders.Key(), B(ws.Authed())):
					// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
				case receivers.Len() == 1 &&
					senders.Len() < 2 &&
					equals(receivers.N(0).Value(), B(ws.Authed())):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
				default:
					// restricted filter: do not return any events,
					//   even if other elements in filters array were not restricted).
					//   client should know better.
					return normalize.Restricted.F("authenticated user %s does not have"+
						" authorization for requested filters", ws.Authed())
				}
			}
		}
		var events []*event.T
		events, err = sto.QueryEvents(c, f)
		if err != nil {
			log.E.F("eventstore: %v", err)
			continue
		}

		// sort in reverse chronological order
		sort.Slice(events, func(i, j int) bool {
			return events[i].CreatedAt.Int() > events[j].CreatedAt.Int()
		})
		for _, ev := range events {
			if s.options.skipEventFunc != nil && s.options.skipEventFunc(ev) {
				continue
			}
			i--
			if i < 0 {
				break
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(env.Subscription.T, ev); chk.E(err) {
				return
			}
			if err = res.Write(ws); chk.E(err) {
				return
			}
		}
	}
	if err = eoseenvelope.NewFrom(env.Subscription).Write(ws); chk.E(err) {
		return
	}
	setListener(env.Subscription.String(), ws, env.Filters)
	return
}

func (s *Server) doClose(c Ctx, ws *web.Socket, req B, store store.I) (note B) {

	var err E
	var rem B
	env := closeenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return B(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.ID.String() == "" {
		return B("CLOSE has no <id>")
	}
	removeListenerId(ws, env.ID.String())
	return
}

func (s *Server) doAuth(c Ctx, ws *web.Socket, req B, store store.I) (msg B) {
	if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
		svcUrl := auther.ServiceUrl(ws.Req())
		if svcUrl == "" {
			return
		}
		log.I.F("received auth response,%s", req)
		var err E
		var rem B
		env := authenvelope.NewResponse()
		if rem, err = env.UnmarshalJSON(req); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.F("extra '%s'", rem)
		}

		var valid bool
		if valid, err = auth.Validate(env.Event, B(ws.Challenge()), svcUrl); chk.E(err) {

			if err := okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F(err.Error())).Write(ws); chk.E(err) {

				return B(err.Error())
			}
			return normalize.Error.F(err.Error())

		} else if !valid {

			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F("failed to authenticate")).Write(ws); chk.E(err) {

				return B(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.ID, true, B{}).Write(ws); chk.E(err) {
				return
			}
			log.I.F("%s authed to pubkey,%0x", ws.RealRemote(), env.Event.PubKey)
			ws.SetAuthed(web.S(env.Event.PubKey))
		}
	}
	return
}

func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E.F("failed to upgrade websocket: %v", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[conn] = struct{}{}
	ticker := time.NewTicker(pingPeriod)

	ip := conn.RemoteAddr().String()
	if realIP := r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP // possible to be multiple comma separated
	} else if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.I.F("connected from %s", ip)
	ws := challenge(conn, r, ip)
	if s.options.perConnectionLimiter != nil {
		ws.SetLimiter(rate.NewLimiter(
			s.options.perConnectionLimiter.Limit(),
			s.options.perConnectionLimiter.Burst(),
		))
	}

	ctx, cancel := context.WithCancel(context.Background())

	store := s.relay.Storage(ctx)

	// reader
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			s.clientsMu.Lock()
			if _, ok := s.clients[conn]; ok {
				conn.Close()
				delete(s.clients, conn)
				removeListener(ws)
			}
			s.clientsMu.Unlock()
			log.I.F("disconnected from %s", ip)
		}()

		conn.SetReadLimit(maxMessageSize)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(S) E {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		if ws.AuthRequested() && len(ws.Authed()) == 0 {
			log.I.F("requesting auth from client from %s", ws.RealRemote())
			if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
				return
			}
			// ignore requests until request is responded to
			return
		}

		for {
			typ, message, err := conn.ReadMessage()
			if chk.E(err) {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,        // 1001
					websocket.CloseNoStatusReceived, // 1005
					websocket.CloseAbnormalClosure,  // 1006
				) {
					log.W.F("unexpected close error from %s: %v",
						r.Header.Get("X-Forwarded-For"), err)
				}
				break
			}

			if ws.Limiter() != nil {
				// NOTE: Wait will throttle the requests.
				// To reject requests exceeding the limit, use if !ws.limiter.Allow()
				if err := ws.Limiter().Wait(context.TODO()); err != nil {
					log.W.F("unexpected limiter error %v", err)
					continue
				}
			}

			if typ == websocket.PingMessage {
				if err = ws.WriteMessage(websocket.PongMessage, nil); chk.E(err) {
					// probably should abort if error here?
				}
				continue
			}

			go s.handleMessage(ctx, ws, message, store)
		}
	}()

	// writer
	go func() {
		defer func() {
			cancel()
			ticker.Stop()
			conn.Close()
		}()
		var err E
		for {
			select {
			case <-ticker.C:
				err = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
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
