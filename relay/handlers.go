package relay

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
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
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

func challenge(conn *websocket.Conn, req *http.Request, addr string) (ws *WebSocket) {
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
	ws = &WebSocket{conn: conn, req: req}
	ws.Remote.Store(addr)
	ws.challenge.Store(S(encoded))
	return
}

func (s *Server) handleMessage(c Ctx, ws *WebSocket, msg B, store store.I) {
	var notice B
	var err E
	defer func() {
		if len(notice) != 0 {
			if err = noticeenvelope.NewFrom(notice).Write(ws); chk.E(err) {
			}
			// ws.WriteJSON(nostr.NoticeEnvelope(notice))
		}
	}()

	var t S
	var rem B
	if t, rem, err = envelopes.Identify(msg); chk.E(err) {
		notice = B(err.Error())
	}

	switch t {
	case eventenvelope.L:
		notice = s.doEvent(c, ws, rem, store)
	case countenvelope.L:
		notice = s.doCount(c, ws, rem, store)
	case reqenvelope.L:
		log.I.F("%s", rem)
		notice = s.doReq(c, ws, rem, store)
	case closeenvelope.L:
		notice = s.doClose(c, ws, rem, store)
	case authenvelope.L:
		notice = s.doAuth(c, ws, rem, store)
	default:
		if cwh, ok := s.relay.(CustomWebSocketHandler); ok {
			cwh.HandleUnknownType(ws, t, rem)
		} else {
			notice = B(fmt.Sprintf("unknown envelope type %s\n%s", t, rem))
		}
	}
	if len(notice) > 0 {
		log.D.F("notice %s", notice)
	}
}

func (s *Server) doEvent(c Ctx, ws *WebSocket, req B, sto store.I) (msg B) {
	var err E
	var ok bool
	var rem B
	advancedDeleter, _ := sto.(AdvancedDeleter)

	env := eventenvelope.NewSubmission()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}

	// latestIndex := len(req) - 1
	//
	// // it's a new event
	// var evt event.T
	// if err := json.Unmarshal(req[latestIndex], &evt); err != nil {
	// 	return B("failed to decode event: " + err.Error())
	// }

	// check id
	if !equals(env.GetIDBytes(), env.ID) {
		// }
		// hash := sha256.Sum256(evt.Serialize())
		// if id := hex.EncodeToString(hash[:]); id != evt.ID {
		// reason := "invalid: event id is computed incorrectly"
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F("event id is computed incorrectly")).Write(ws); chk.E(err) {
			return
		}
		// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false, Reason: reason})
		return
	}

	// check signature
	if ok, err = env.Verify(); err != nil {

		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("failed to verify signature")).Write(ws); chk.E(err) {
			return
		}

		// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false,
		// 	Reason: "error: failed to verify signature"})
		// return ""
	} else if !ok {

		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Error.F("signature is invalid")).Write(ws); chk.E(err) {
			return
		}

		// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false,
		// 	Reason: "invalid: signature is invalid"})
		// return ""
		return
	}
	// log.I.F("%v %s %v %v", env.T.Kind, kind.GetString(env.T.Kind), env.Kind.K, kind.Deletion.K)
	if env.T.Kind.K == kind.Deletion.K {
		log.I.F("delete event\n%s", env.T.Serialize())
		// event deletion -- nip09
		// log.I.S(env.Tags, env.Tags.Value())
		for _, t := range env.Tags.Value() {
			var res []*event.T
			if t.Len() >= 2 {
				switch {
				case equals(t.Key(), B("e")):
					// fetch event to be deleted
					evId := make(B, sha256.Size)
					if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
						continue
					}
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
					// log.I.S(split)
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
					// log.I.S(f)
					res, err = s.relay.Storage(c).QueryEvents(c, f)
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
					log.I.S(res)
				}
			}

			// var target *event.T
			// exists := false
			// select {
			// case target, exists = <-res:
			// case <-ctx.Done():
			// }
			if len(res) < 1 {
				// this will happen if event is not in the database
				// or when when the query is taking too long, so we just give up
				continue
			}
			// log.I.S(res)
			// there can only be one
			// target := res[0]
			for _, target := range res {
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
					// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false,
					// 	Reason: "insufficient permissions"})
					return
				}

				if advancedDeleter != nil {
					advancedDeleter.BeforeDelete(c, t.Value(), env.PubKey)
				}

				if err = sto.DeleteEvent(c, target.EventID()); err != nil {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
						return
					}

					// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: false,
					// 	Reason: fmt.Sprintf("error: %s", err.Error())})
					return
				}

				if advancedDeleter != nil {
					advancedDeleter.AfterDelete(t.Value(), env.PubKey)
				}
			}
		}

		notifyListeners(env.T)
		if err = okenvelope.NewFrom(env.ID, true).Write(ws); chk.E(err) {
			return
		}
		// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: true})
		return
	}

	ok, reason := AddEvent(c, s.relay, env.T)
	if err = okenvelope.NewFrom(env.ID, ok, reason).Write(ws); chk.E(err) {
		return
	}
	// ws.WriteJSON(nostr.OKEnvelope{EventID: evt.ID, OK: ok, Reason: reason})
	return
}

func (s *Server) doCount(c context.Context, ws *WebSocket, req B,
	store store.I) (msg B) {

	counter, ok := store.(EventCounter)
	if !ok {
		return normalize.Restricted.F("this relay does not support NIP-45")
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

	// var id S
	// json.Unmarshal(req[1], &id)
	if env.ID == nil || env.ID.String() == "" {
		return normalize.Error.F("COUNT has no <id>")
	}

	var total N
	// ff := make(nostr.Filters, len(req)-2)
	// ff := filters.T{F: make([]*filter.T, len(req)-2)}
	// for i, filterReq := range req[2:] {
	for _, f := range env.Filters.F {
		// if err := json.Unmarshal(filterReq, &ff.F[i]); err != nil {
		// 	return normalize.Error.F("failed to decode filter")
		// }

		// prevent kind-4 events from being returned to unauthed users,
		//   only when authentication is a thing
		if _, ok = s.relay.(Authenticator); ok {
			if f.Kinds.Contains(kind.EncryptedDirectMessage) {
				// if slices.Contains(f.Kinds.K, kind.EncryptedDirectMessage) {
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("p"))
				switch {
				case len(ws.authed) == 0:
					// not authenticated
					return normalize.Restricted.F(
						"this relay does not serve kind-4 to unauthenticated users," +
							" does your client implement NIP-42?")
				case senders.Len() == 1 &&
					receivers.Len() < 2 &&
					equals(senders.F()[0], ws.authed):
					// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
				case receivers.Len() == 1 &&
					senders.Len() < 2 &&
					equals(receivers.N(0).Value(), ws.authed):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
				default:
					// restricted filter: do not return any events,
					//   even if other elements in filters array were not restricted).
					//   client should know better.
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
	if err = countenvelope.NewResponseFrom(env.ID.String(), N(total),
		false).Write(ws); chk.E(err) {
		return
	}
	// ws.WriteJSON([]interface{}{"COUNT", id, map[S]int64{"count": total}})
	return
}

func (s *Server) doReq(c Ctx, ws *WebSocket, req B, sto store.I) (r B) {

	var err E
	var rem B
	env := reqenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}

	if accepter, ok := s.relay.(ReqAcceptor); ok {
		if !accepter.AcceptReq(c, env.Subscription.T, env.Filters, ws.authed) {
			return B("REQ filters are not accepted")
		}
	}

	// for _, f := range ff {
	for _, f := range env.Filters.F {
		var i uint
		if filter.Present(f.Limit) {
			if *f.Limit == 0 {
				// log.I.F("filter explicitly zero %s\n%s", env.Subscription.String(),
				// 	f.String())
				continue
			}
			i = *f.Limit
		}
		// prevent kind-4 events from being returned to unauthed users,
		//   only when authentication is a thing
		if _, ok := s.relay.(Authenticator); ok {
			if f.Kinds.Contains(kind.EncryptedDirectMessage) {
				// if slices.Contains(f.Kinds.K, kind.EncryptedDirectMessage) {
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("p"))
				switch {
				case len(ws.authed) == 0:
					// not authenticated
					return normalize.Restricted.F(
						"this relay does not serve kind-4 to unauthenticated users," +
							" does your client implement NIP-42?")
				case senders.Len() == 1 &&
					receivers.Len() < 2 &&
					equals(senders.Key(), ws.authed):
					// allowed filter: ws.authed is sole sender (filter specifies one or all receivers)
				case receivers.Len() == 1 &&
					senders.Len() < 2 &&
					equals(receivers.N(0).Value(), ws.authed):
					// allowed filter: ws.authed is sole receiver (filter specifies one or all senders)
				default:
					// restricted filter: do not return any events,
					//   even if other elements in filters array were not restricted).
					//   client should know better.
					return normalize.Restricted.F("authenticated user does not have" +
						" authorization for requested filters")
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
			if err = eventenvelope.NewResultWith(env.Subscription.T, ev).Write(ws); chk.E(err) {
				continue
			}
			// ws.WriteJSON(nostr.EventEnvelope{SubscriptionID: &id, Event: *ev})
		}

		// // exhaust the channel (in case we broke out of it early) so it is closed by the storage
		// for range events {
		// }
	}
	if err = eoseenvelope.NewFrom(env.Subscription).Write(ws); chk.E(err) {
		return
	}
	// ws.WriteJSON(nostr.EOSEEnvelope(id))
	setListener(env.Subscription.String(), ws, env.Filters)
	return
}

func (s *Server) doClose(c Ctx, ws *WebSocket, req B, store store.I) (note B) {

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

func (s *Server) doAuth(c Ctx, ws *WebSocket, req B, store store.I) (msg B) {
	if auther, ok := s.relay.(Authenticator); ok {
		log.D.F("received auth response\n%s", req)
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
		if valid, err = auth.Validate(env.Event, auth.B(ws.challenge.Load()),
			auther.ServiceUrl(ws.req)); chk.E(err) {
			if err = okenvelope.NewFrom(env.Event.ID, false, normalize.Error.F(err.Error())).
				Write(ws); chk.E(err) {
				return B(err.Error())
			}
			return normalize.Error.F(err.Error())
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F("failed to authenticate")).
				Write(ws); chk.E(err) {
				return B(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			log.D.F("%s authed to pubkey %0x", ws.Remote.Load(),
				env.Event.PubKey)
			ws.authed = env.Event.PubKey
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
		ws.limiter = rate.NewLimiter(
			s.options.perConnectionLimiter.Limit(),
			s.options.perConnectionLimiter.Burst(),
		)
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

		// NIP-42 auth challenge
		if _, ok := s.relay.(Authenticator); ok {
			env := authenvelope.NewChallengeWith(ws.challenge.String())
			buf := bytes.NewBuffer(nil)
			env.Write(buf)
			log.I.F("requesting auth '%s", buf.String())
			if err = env.Write(ws); chk.E(err) {
				return
			}
		}

		for {
			typ, message, err := conn.ReadMessage()
			if err != nil {
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

			if ws.limiter != nil {
				// NOTE: Wait will throttle the requests.
				// To reject requests exceeding the limit, use if !ws.limiter.Allow()
				if err := ws.limiter.Wait(context.TODO()); err != nil {
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
				log.I.F("pinging for %s", ip)
			case <-ctx.Done():
				return
			}
		}
	}()
}
