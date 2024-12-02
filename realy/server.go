package realy

import (
	"bytes"
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/fasthttp/websocket"
	"github.com/rs/cors"
	"golang.org/x/time/rate"

	"realy.lol/auth"
	"realy.lol/cmd/realy/app"
	"realy.lol/context"
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
	"realy.lol/kinds"
	"realy.lol/normalize"
	"realy.lol/realy/options"
	"realy.lol/relay"
	"realy.lol/relay/wrapper"
	"realy.lol/relayinfo"
	"realy.lol/sha256"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/web"
)

type Server struct {
	Ctx                  cx
	Cancel               context.F
	options              *options.T
	relay                relay.I
	clientsMu            sync.Mutex
	clients              map[*websocket.Conn]struct{}
	Addr                 st
	serveMux             *http.ServeMux
	httpServer           *http.Server
	authRequired         bo
	maxLimit             no
	adminUser, adminPass st
}

type ServerParams struct {
	Ctx                  cx
	Cancel               context.F
	Rl                   relay.I
	DbPath               st
	MaxLimit             no
	AdminUser, AdminPass st
}

func NewServer(sp ServerParams, opts ...options.O) (*Server, er) {
	op := options.Default()
	for _, opt := range opts {
		opt(op)
	}
	var authRequired bo
	if ar, ok := sp.Rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	srv := &Server{Ctx: sp.Ctx, Cancel: sp.Cancel, relay: sp.Rl,
		clients: make(map[*websocket.Conn]struct{}), serveMux: http.NewServeMux(),
		options: op, authRequired: authRequired, maxLimit: sp.MaxLimit,
		adminUser: sp.AdminUser, adminPass: sp.AdminPass}
	if storage := sp.Rl.Storage(context.Bg()); storage != nil {
		if err := storage.Init(sp.DbPath); chk.T(err) {
			return nil, fmt.Errorf("storage init: %w", err)
		}
	}
	if err := sp.Rl.Init(); chk.T(err) {
		return nil, fmt.Errorf("realy init: %w", err)
	}
	if inj, ok := sp.Rl.(relay.Injector); ok {
		go func() {
			for ev := range inj.InjectEvents() {
				notifyListeners(srv.authRequired, ev)
			}
		}()
	}
	return srv, nil
}

func (s *Server) HTTPAuth(r *http.Request) (authed bo) {
	if s.adminUser == "" || s.adminPass == "" {
		// disallow this if it hasn't been configured, the default values are empty.
		return
	}
	username, password, ok := r.BasicAuth()
	if ok {
		usernameHash := sha256.Sum256(by(username))
		passwordHash := sha256.Sum256(by(password))
		expectedUsernameHash := sha256.Sum256(by(s.adminUser))
		expectedPasswordHash := sha256.Sum256(by(s.adminPass))
		usernameMatch := subtle.ConstantTimeCompare(usernameHash[:],
			expectedUsernameHash[:]) == 1
		passwordMatch := subtle.ConstantTimeCompare(passwordHash[:],
			expectedPasswordHash[:]) == 1
		if usernameMatch && passwordMatch {
			return true
		}
	}
	return
}

func (s *Server) AuthFail(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	fmt.Fprintf(w, "you may have not configured your admin username/password")
}

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/export"):
		if ok := s.HTTPAuth(r); !ok {
			s.AuthFail(w)
			return
		}
		log.I.F("export of event data requested on admin port")
		sto := s.relay.Storage(context.Bg())
		if strings.Count(r.URL.Path, "/") > 1 {
			split := strings.Split(r.URL.Path, "/")
			if len(split) != 3 {
				fprintf(w, "incorrectly formatted export parameter: '%s'", r.URL.Path)
				return
			}
			switch split[2] {
			case "users":
				if rl, ok := s.relay.(*app.Relay); ok {
					follows := make([]by, 0, len(rl.Followed))
					for f := range rl.Followed {
						follows = append(follows, by(f))
					}
					sto.Export(s.Ctx, w, follows...)
				}
			default:
				var exportPubkeys []by
				pubkeys := strings.Split(split[2], "-")
				for _, pubkey := range pubkeys {
					pk, err := hex.Dec(pubkey)
					if err != nil {
						log.E.F("invalid public key '%s' in parameters", pubkey)
						continue
					}
					exportPubkeys = append(exportPubkeys, pk)
				}
				sto.Export(s.Ctx, w, exportPubkeys...)
			}
		} else {
			sto.Export(s.Ctx, w)
		}
	case strings.HasPrefix(r.URL.Path, "/import"):
		if ok := s.HTTPAuth(r); !ok {
			s.AuthFail(w)
			return
		}
		log.I.F("import of event data requested on admin port %s", r.RequestURI)
		sto := s.relay.Storage(context.Bg())
		read := io.LimitReader(r.Body, r.ContentLength)
		sto.Import(read)
	case strings.HasPrefix(r.URL.Path, "/shutdown"):
		if ok := s.HTTPAuth(r); !ok {
			s.AuthFail(w)
			return
		}
		fprintf(w, "shutting down")
		defer chk.E(r.Body.Close())
		s.Shutdown()
	default:
		fprintf(w, "todo: realy web interface page\n\n")
		s.HandleNIP11(w, r)
	}
}

func (s *Server) HandleNIP11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.T.Ln("handling relay information document")
	var info *relayinfo.T
	if informationer, ok := s.relay.(relay.Informationer); ok {
		info = informationer.GetNIP11InformationDocument()
	} else {
		supportedNIPs := relayinfo.GetList(relayinfo.BasicProtocol, relayinfo.EventDeletion,
			relayinfo.RelayInformationDocument, relayinfo.GenericTagQueries,
			relayinfo.NostrMarketplace,
			relayinfo.EventTreatment, relayinfo.CommandResults,
			relayinfo.ParameterizedReplaceableEvents,
			relayinfo.ProtectedEvents)
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok && auther.ServiceUrl(r) != "" {
			supportedNIPs = append(supportedNIPs, relayinfo.Authentication.N())
		}
		var storage store.I
		if s.relay.Storage(context.Bg()) != nil {
			if _, ok = storage.(relay.EventCounter); ok {
				supportedNIPs = append(supportedNIPs, relayinfo.CountingResults.N())
			}
		}
		log.T.Ln("supported NIPs", supportedNIPs)
		info = &relayinfo.T{Name: s.relay.Name(),
			Description: "relay powered by the realy framework",
			Nips:        supportedNIPs, Software: "https://realy.lol", Version: version,
			Limitation: relayinfo.Limits{MaxLimit: s.maxLimit, AuthRequired: s.authRequired},
			Icon:       "https://cdn.satellite.earth/ac9778868fbf23b63c47c769a74e163377e6ea94d3f0f31711931663d035c4f6.png"}
	}
	if err := json.NewEncoder(w).Encode(info); chk.E(err) {
	}
}

func addEvent(c cx, rl relay.I, ev *event.T, hr *http.Request, origin st,
	authedPubkey by) (accepted bo, message by) {
	if ev == nil {
		return false, normalize.Invalid.F("empty event")
	}
	sto := rl.Storage(c)
	wrap := &wrapper.Relay{I: sto}
	advancedSaver, _ := sto.(relay.AdvancedSaver)
	accept, notice, after := rl.AcceptEvent(c, ev, hr, origin, authedPubkey)
	if !accept {
		return false, normalize.Blocked.F(notice)
	}
	if ev.Tags.ContainsProtectedMarker() {
		if len(authedPubkey) == 0 || !equals(ev.PubKey, authedPubkey) {
			return false, by(fmt.Sprintf("event with relay marker tag '-' may only be published by matching npub: %0x is not %0x",
				authedPubkey, ev.PubKey))
		}
	}
	if ev.Kind.IsEphemeral() {
	} else {
		if advancedSaver != nil {
			advancedSaver.BeforeSave(c, ev)
		}
		if saveErr := wrap.Publish(c, ev); chk.E(saveErr) {
			if errors.Is(saveErr, store.ErrDupEvent) {
				return false, normalize.Error.F(saveErr.Error())
			}
			errmsg := saveErr.Error()
			if nip20prefixmatcher.MatchString(errmsg) {
				if strings.Contains(errmsg, "tombstone") {
					return false, normalize.Blocked.F("event was deleted, not storing it again")
				}
				return false, normalize.Error.F(errmsg)
			} else {
				return false, normalize.Error.F("failed to save (%s)", errmsg)
			}
		}
		if advancedSaver != nil {
			advancedSaver.AfterSave(ev)
		}
	}
	var authRequired bo
	if ar, ok := rl.(relay.Authenticator); ok {
		authRequired = ar.AuthEnabled()
	}
	if after != nil {
		after()
	}
	notifyListeners(authRequired, ev)
	accepted = true
	return
}

func (s *Server) doEvent(c cx, ws *web.Socket, req by, sto store.I) (msg by) {
	log.T.F("doEvent %s %s", ws.RealRemote(), req)
	var err er
	var ok bo
	var rem by
	advancedDeleter, _ := sto.(relay.AdvancedDeleter)
	env := eventenvelope.NewSubmission()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	accept, notice, after := s.relay.AcceptEvent(c, env.T, ws.Req(), ws.RealRemote(),
		by(ws.Authed()))
	if !accept {
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if !ws.AuthRequested() {
				if err = okenvelope.NewFrom(env.ID, false,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.T(err) {
				}
				log.T.F("requesting auth from client %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.T(err) {
					return
				}
				ws.RequestAuth()
				return
			} else {
				if err = okenvelope.NewFrom(env.ID, false,
					normalize.AuthRequired.F("auth required for storing events")).Write(ws); chk.T(err) {
				}
				log.T.F("requesting auth again from client %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.T(err) {
					return
				}
				return
			}
		}
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F(notice)).Write(ws); chk.T(err) {
		}
		return
	}
	if !equals(env.GetIDBytes(), env.ID) {
		if err = okenvelope.NewFrom(env.ID, false,
			normalize.Invalid.F("event id is computed incorrectly")).Write(ws); chk.E(err) {
			return
		}
		return
	}
	if ok, err = env.Verify(); chk.T(err) {
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
		for _, t := range env.Tags.Value() {
			var res []*event.T
			if t.Len() >= 2 {
				switch {
				case equals(t.Key(), by("e")):
					evId := make(by, sha256.Size)
					if _, err = hex.DecBytes(evId, t.Value()); chk.E(err) {
						continue
					}
					res, err = s.relay.Storage(c).QueryEvents(c, &filter.T{IDs: tag.New(evId)})
					if err != nil {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("failed to query for target event")).Write(ws); chk.E(err) {
							return
						}
						return
					}
					for i := range res {
						if res[i].Kind.Equal(kind.Deletion) {
							if err = okenvelope.NewFrom(env.ID, false,
								normalize.Blocked.F("not processing or storing delete event containing delete event references")).Write(ws); chk.E(err) {
								return
							}
						}
						if !equals(res[i].PubKey, env.T.PubKey) {
							if err = okenvelope.NewFrom(env.ID, false,
								normalize.Blocked.F("cannot delete other users' events")).Write(ws); chk.E(err) {
								return
							}
						}
					}
				case equals(t.Key(), by("a")):
					split := bytes.Split(t.Value(), by{':'})
					if len(split) != 3 {
						continue
					}
					kin := ints.New(uint16(0))
					if _, err = kin.UnmarshalJSON(split[0]); chk.E(err) {
						return
					}
					kk := kind.New(kin.Uint16())
					if kk.Equal(kind.Deletion) {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Blocked.F("delete event kind may not be deleted")).Write(ws); chk.E(err) {
							return
						}
					}
					if !kk.IsParameterizedReplaceable() {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Error.F("delete tags with a tags containing non-parameterized-replaceable events cannot be processed")).Write(ws); chk.E(err) {
							return
						}
					}
					if !equals(split[1], env.T.PubKey) {
						if err = okenvelope.NewFrom(env.ID, false,
							normalize.Blocked.F("cannot delete other users' events")).Write(ws); chk.E(err) {
							return
						}
					}
					f := filter.New()
					f.Kinds.K = []*kind.T{kk}
					aut := make(by, 0, len(split[1])/2)
					if aut, err = hex.DecAppend(aut, split[1]); chk.E(err) {
						return
					}
					f.Authors.Append(aut)
					f.Tags.AppendTags(tag.New(by{'#', 'd'}, split[2]))
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
				continue
			}
			var resTmp []*event.T
			for _, v := range res {
				if env.T.CreatedAt.U64() >= v.CreatedAt.U64() {
					resTmp = append(resTmp, v)
				}
			}
			res = resTmp
			for _, target := range res {
				if target.Kind.K == kind.Deletion.K {
					if err = okenvelope.NewFrom(env.ID, false,
						normalize.Error.F("cannot delete delete event %s",
							env.ID)).Write(ws); chk.E(err) {
						return
					}
				}
				if target.CreatedAt.Int() > env.T.CreatedAt.Int() {
					log.I.F("not deleting\n%d%\nbecause delete event is older\n%d",
						target.CreatedAt.Int(), env.T.CreatedAt.Int())
					continue
				}
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
				if err = sto.DeleteEvent(c, target.EventID()); chk.T(err) {
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
			res = nil
		}
		if err = okenvelope.NewFrom(env.ID, true).Write(ws); chk.E(err) {
			return
		}
	}
	ok, reason := addEvent(c, s.relay, env.T, ws.Req(), ws.RealRemote(), by(ws.Authed()))
	if err = okenvelope.NewFrom(env.ID, ok, reason).Write(ws); chk.E(err) {
		return
	}
	if after != nil {
		after()
	}
	return
}

func (s *Server) doCount(c context.T, ws *web.Socket, req by, store store.I) (msg by) {
	counter, ok := store.(relay.EventCounter)
	if !ok {
		return normalize.Restricted.F("this relay does not support NIP-45")
	}
	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		return
	}
	var err er
	var rem by
	env := countenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.Subscription == nil || env.Subscription.String() == "" {
		return normalize.Error.F("COUNT has no <subscription id>")
	}
	allowed := env.Filters
	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		var accepted bo
		allowed, accepted = accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters,
			by(ws.Authed()))
		if !accepted || allowed == nil {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for count processing")).Write(ws); chk.E(err) {
				}
				log.I.F("requesting auth from client from %s", ws.RealRemote())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}
	}
	if allowed != env.Filters {
		defer func() {
			var auther relay.Authenticator
			var ok bo
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
				}
				log.T.F("requesting auth from client from %s, challenge '%s'", ws.RealRemote(),
					ws.Challenge())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}()
	}
	var total no
	var approx bo
	if allowed != nil {
		for _, f := range allowed.F {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
				if f.Kinds.Contains(kind.EncryptedDirectMessage) || f.Kinds.Contains(kind.GiftWrap) {
					senders := f.Authors
					receivers := f.Tags.GetAll(tag.New("p"))
					switch {
					case len(ws.Authed()) == 0:
						return normalize.Restricted.F("this realy does not serve kind-4 to unauthenticated users," + " does your client implement NIP-42?")
					case senders.Len() == 1 && receivers.Len() < 2 && equals(senders.F()[0],
						by(ws.Authed())):
					case receivers.Len() == 1 && senders.Len() < 2 && equals(receivers.N(0).Value(),
						by(ws.Authed())):
					default:
						return normalize.Restricted.F("authenticated user does not have" + " authorization for requested filters")
					}
				}
			}
			var count no
			count, approx, err = counter.CountEvents(c, f)
			if err != nil {
				log.E.F("store: %v", err)
				continue
			}
			total += count
		}
	}
	var res *countenvelope.Response
	if res, err = countenvelope.NewResponseFrom(env.Subscription.T, total, approx); chk.E(err) {
		return
	}
	if err = res.Write(ws); chk.E(err) {
		return
	}
	return
}

func (s *Server) doReq(c cx, ws *web.Socket, req by, sto store.I) (r by) {
	if ws.AuthRequested() && len(ws.Authed()) == 0 {
		return
	}
	var err er
	var rem by
	env := reqenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return normalize.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	allowed := env.Filters
	if accepter, ok := s.relay.(relay.ReqAcceptor); ok {
		var accepted bo
		allowed, accepted = accepter.AcceptReq(c, ws.Req(), env.Subscription.T, env.Filters,
			by(ws.Authed()))
		if !accepted || allowed == nil {
			var auther relay.Authenticator
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
				}
				log.T.F("requesting auth from client from %s, challenge '%s'", ws.RealRemote(),
					ws.Challenge())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
			return
		}
	}
	if allowed != env.Filters {
		defer func() {
			var auther relay.Authenticator
			var ok bo
			if auther, ok = s.relay.(relay.Authenticator); ok && auther.AuthEnabled() && !ws.AuthRequested() {
				ws.RequestAuth()
				if err = closedenvelope.NewFrom(env.Subscription,
					normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
				}
				log.T.F("requesting auth from client from %s, challenge '%s'", ws.RealRemote(),
					ws.Challenge())
				if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
					return
				}
				return
			}
		}()
	}
	for _, f := range allowed.F {
		var i uint
		if filter.Present(f.Limit) {
			if *f.Limit == 0 {
				continue
			}
			i = *f.Limit
		}
		if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
			if f.Kinds.IsPrivileged() {
				log.T.F("privileged request with auth enabled\n%s", f.Serialize())
				senders := f.Authors
				receivers := f.Tags.GetAll(tag.New("#p"))
				switch {
				case len(ws.Authed()) == 0:
					ws.RequestAuth()
					if err = closedenvelope.NewFrom(env.Subscription,
						normalize.AuthRequired.F("auth required for request processing")).Write(ws); chk.E(err) {
					}
					log.I.F("requesting auth from client from %s", ws.RealRemote())
					if err = authenvelope.NewChallengeWith(ws.Challenge()).Write(ws); chk.E(err) {
						return
					}
					notice := normalize.Restricted.F("this realy does not serve kind-4 to unauthenticated users," + " does your client implement NIP-42?")
					return notice
				case senders.Contains(ws.AuthedBytes()) || receivers.ContainsAny(by("#p"),
					tag.New(ws.AuthedBytes())):
					log.T.F("user %0x from %s allowed to query for privileged event",
						ws.AuthedBytes(), ws.RealRemote())
				default:
					return normalize.Restricted.F("authenticated user %0x does not have"+" authorization for requested filters",
						ws.AuthedBytes())
				}
			}
		}
		var events event.Ts
		log.D.F("query from %s %0x,%s", ws.RealRemote(), ws.AuthedBytes(), f.Serialize())
		if events, err = sto.QueryEvents(c, f); err != nil {
			log.E.F("eventstore: %v", err)
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
			continue
		}
		if aut := ws.Authed(); ws.IsAuthed() {
			var mutes event.Ts
			if mutes, err = sto.QueryEvents(c, &filter.T{Authors: tag.New(aut),
				Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
				var mutePubs []by
				for _, ev := range mutes {
					for _, t := range ev.Tags.F() {
						if equals(t.Key(), by("p")) {
							var p by
							if p, err = hex.Dec(st(t.Value())); chk.E(err) {
								continue
							}
							mutePubs = append(mutePubs, p)
						}
					}
				}
				var tmp event.Ts
				for _, ev := range events {
					for _, pk := range mutePubs {
						if equals(ev.PubKey, pk) {
							continue
						}
						tmp = append(tmp, ev)
					}
				}
				events = tmp
			}
		}
		sort.Slice(events, func(i, j int) bo {
			return events[i].CreatedAt.Int() > events[j].CreatedAt.Int()
		})
		for _, ev := range events {
			if s.options.SkipEventFunc != nil && s.options.SkipEventFunc(ev) {
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
	if env.Filters != allowed {
		return
	}
	setListener(env.Subscription.String(), ws, env.Filters)
	return
}

func (s *Server) doClose(ws *web.Socket, req by) (note by) {
	var err er
	var rem by
	env := closeenvelope.New()
	if rem, err = env.UnmarshalJSON(req); chk.E(err) {
		return by(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	if env.ID.String() == "" {
		return by("CLOSE has no <id>")
	}
	removeListenerId(ws, env.ID.String())
	return
}

func (s *Server) doAuth(ws *web.Socket, req by) (msg by) {
	if auther, ok := s.relay.(relay.Authenticator); ok && auther.AuthEnabled() {
		svcUrl := auther.ServiceUrl(ws.Req())
		if svcUrl == "" {
			return
		}
		log.T.F("received auth response,%s", req)
		var err er
		var rem by
		env := authenvelope.NewResponse()
		if rem, err = env.UnmarshalJSON(req); chk.E(err) {
			return
		}
		if len(rem) > 0 {
			log.I.F("extra '%s'", rem)
		}
		var valid bo
		if valid, err = auth.Validate(env.Event, by(ws.Challenge()), svcUrl); chk.E(err) {
			if err := okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F(err.Error())).Write(ws); chk.E(err) {
				return by(err.Error())
			}
			return normalize.Error.F(err.Error())
		} else if !valid {
			if err = okenvelope.NewFrom(env.Event.ID, false,
				normalize.Error.F("failed to authenticate")).Write(ws); chk.E(err) {
				return by(err.Error())
			}
			return normalize.Restricted.F("auth response does not validate")
		} else {
			if err = okenvelope.NewFrom(env.Event.ID, true, by{}).Write(ws); chk.E(err) {
				return
			}
			log.D.F("%s authed to pubkey,%0x", ws.RealRemote(), env.Event.PubKey)
			ws.SetAuthed(st(env.Event.PubKey))
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
	var realIP st
	if realIP = r.Header.Get("X-Forwarded-For"); realIP != "" {
		ip = realIP
	} else if realIP = r.Header.Get("X-Real-Ip"); realIP != "" {
		ip = realIP
	}
	log.T.F("connected from %s", ip)
	ws := challenge(conn, r, ip)
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
				removeListener(ws)
			}
			s.clientsMu.Unlock()
		}()
		conn.SetReadLimit(maxMessageSize)
		chk.E(conn.SetReadDeadline(time.Now().Add(pongWait)))
		conn.SetPongHandler(func(st) er {
			chk.E(conn.SetReadDeadline(time.Now().Add(pongWait)))
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
		notice = s.doEvent(c, ws, rem, sto)
	case countenvelope.L:
		notice = s.doCount(c, ws, rem, sto)
	case reqenvelope.L:
		notice = s.doReq(c, ws, rem, sto)
	case closeenvelope.L:
		notice = s.doClose(ws, rem)
	case authenvelope.L:
		notice = s.doAuth(ws, rem)
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		s.HandleWebsocket(w, r)
		return
	} else if r.Header.Get("Accept") == "application/nostr+json" {
		s.HandleNIP11(w, r)
		return
	}
	s.HandleAdmin(w, r)
}

func (s *Server) Start(host st, port int, started ...chan bo) er {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	log.I.F("starting relay listener at %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Addr = ln.Addr().String()
	s.httpServer = &http.Server{Handler: cors.Default().Handler(s), Addr: addr,
		WriteTimeout: 7 * time.Second, ReadTimeout: 7 * time.Second,
		IdleTimeout: 28 * time.Second}
	for _, startedC := range started {
		close(startedC)
	}
	if err = s.httpServer.Serve(ln); errors.Is(err, http.ErrServerClosed) {
	} else if err != nil {
	}
	return nil
}

func (s *Server) Shutdown() {
	log.I.Ln("shutting down relay")
	s.Cancel()
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		log.I.Ln("disconnecting", conn.RemoteAddr())
		chk.E(conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(time.Second)))
		chk.E(conn.Close())
		delete(s.clients, conn)
	}
	log.W.Ln("closing event store")
	chk.E(s.relay.Storage(s.Ctx).Close())
	log.W.Ln("shutting down relay listener")
	chk.E(s.httpServer.Shutdown(s.Ctx))
	if f, ok := s.relay.(relay.ShutdownAware); ok {
		f.OnShutdown(s.Ctx)
	}
}

func (s *Server) Router() *http.ServeMux {
	return s.serveMux
}

func fprintf(w io.Writer, format st, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
