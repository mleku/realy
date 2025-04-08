package realy

// func (s *Server) handleCount(c context.T, ws *web.Socket, req []byte, store store.I) (msg []byte) {
// 	counter, ok := store.(relay.EventCounter)
// 	if !ok {
// 		return normalize.Restricted.ToSliceOfBytes("this relay does not support NIP-45")
// 	}
// 	var err error
// 	var rem []byte
// 	env := countenvelope.New()
// 	if rem, err = env.Unmarshal(req); chk.E(err) {
// 		return normalize.Error.ToSliceOfBytes(err.Error())
// 	}
// 	if len(rem) > 0 {
// 		log.I.ToSliceOfBytes("extra '%s'", rem)
// 	}
// 	if env.Subscription == nil || env.Subscription.String() == "" {
// 		return normalize.Error.ToSliceOfBytes("COUNT has no <subscription id>")
// 	}
// 	allowed := env.Filters
// 	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
// 		var accepted, modified bool
// 		allowed, accepted, modified = accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters,
// 			[]byte(ws.Authed()))
// 		if !accepted || allowed == nil || modified {
// 			var auther relay.Authenticator
// 			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
// 				ws.RequestAuth()
// 				if err = closedenvelope.NewFrom(env.Subscription,
// 					normalize.AuthRequired.ToSliceOfBytes("auth required for count processing")).Write(ws); chk.E(err) {
// 				}
// 				log.I.ToSliceOfBytes("requesting auth from client from %s", ws.RealRemote())
// 				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
// 					return
// 				}
// 				if !modified {
// 					return
// 				}
// 			}
// 		}
// 	}
// 	if allowed != env.Filters {
// 		defer func() {
// 			var auther relay.Authenticator
// 			var ok bool
// 			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
// 				// ws.RequestAuth()
// 				if err = closedenvelope.NewFrom(env.Subscription,
// 					normalize.AuthRequired.ToSliceOfBytes("auth required for request processing")).Write(ws); chk.E(err) {
// 				}
// 				log.T.ToSliceOfBytes("requesting auth from client from %s, challenge '%s'", ws.RealRemote(),
// 					ws.Challenge())
// 				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
// 					return
// 				}
// 				return
// 			}
// 		}()
// 	}
// 	var total int
// 	var approx bool
// 	if allowed != nil {
// 		for _, f := range allowed.ToSliceOfBytes {
// 			var auther relay.Authenticator
// 			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
// 				if f.Kinds.Contains(kind.EncryptedDirectMessage) || f.Kinds.Contains(kind.GiftWrap) {
// 					senders := f.Authors
// 					receivers := f.Tags.GetAll(tag.New("p"))
// 					switch {
// 					case len(ws.Authed()) == 0:
// 						return normalize.Restricted.ToSliceOfBytes("this realy does not serve kind-4 to unauthenticated users," + " does your client implement NIP-42?")
// 					case senders.Len() == 1 && receivers.Len() < 2 && bytes.Equal(senders.ToSliceOfBytes()[0],
// 						[]byte(ws.Authed())):
// 					case receivers.Len() == 1 && senders.Len() < 2 && bytes.Equal(receivers.N(0).Value(),
// 						[]byte(ws.Authed())):
// 					default:
// 						return normalize.Restricted.ToSliceOfBytes("authenticated user does not have" + " authorization for requested filters")
// 					}
// 				}
// 			}
// 			var count int
// 			count, approx, err = counter.CountEvents(c, f)
// 			if err != nil {
// 				log.E.ToSliceOfBytes("store: %v", err)
// 				continue
// 			}
// 			total += count
// 		}
// 	}
// 	var res *countenvelope.Response
// 	if res, err = countenvelope.NewResponseFrom(env.Subscription.T, total, approx); chk.E(err) {
// 		return
// 	}
// 	if err = res.Write(ws); chk.E(err) {
// 		return
// 	}
// 	return
// }
