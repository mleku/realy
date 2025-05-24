package realy

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/cors"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/list"
	"realy.lol/log"
	"realy.lol/realy/config"
	"realy.lol/realy/helpers"
	"realy.lol/servemux"
	"realy.lol/signer"
	"realy.lol/store"
)

type Server struct {
	Name       string
	Ctx        context.T
	Cancel     context.F
	WG         *sync.WaitGroup
	Address    string
	HTTPServer *http.Server
	Mux        *servemux.S
	huma.API
	Store    store.I
	MaxLimit int

	configurationMx sync.Mutex
	configuration   config.C

	sync.Mutex
	Superuser signer.I
	admins    []signer.I
	owners    [][]byte
	// followed are the pubkeys that are in the Owners' follow lists and have full
	// access permission.
	followed list.L
	// OwnersFollowed are "guests" of the followed and have full access but with
	// rate limiting enabled.
	ownersFollowed list.L
	// // muted are on Owners' mute lists and do not have write access to the relay,
	// // even if they would be in the OwnersFollowed list, they can only read.
	// muted list.L
	// ownersFollowLists are the event IDs of owners follow lists, which must not be
	// deleted, only replaced.
	ownersFollowLists [][]byte
	// ownersMuteLists are the event IDs of owners mute lists, which must not be
	// deleted, only replaced.
	ownersMuteLists [][]byte
}

func (s *Server) Start() (err error) {
	s.Init()
	var listener net.Listener
	if listener, err = net.Listen("tcp", s.Address); chk.E(err) {
		return
	}
	s.HTTPServer = &http.Server{
		Handler:           cors.Default().Handler(s),
		Addr:              s.Address,
		ReadHeaderTimeout: 7 * time.Second,
		IdleTimeout:       28 * time.Second,
	}
	log.I.F("listening on %s", s.Address)
	if err = s.HTTPServer.Serve(listener); errors.Is(err, http.ErrServerClosed) {
		return
	} else if chk.E(err) {
		return
	}
	return
}

// ServeHTTP is the server http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote := helpers.GetRemoteFromReq(r)
	allowList := s.Configuration().AllowList
	if len(allowList) > 0 {
		var allowed bool
		for _, a := range allowList {
			if strings.HasPrefix(remote, a) {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}
	blocklist := s.Configuration().BlockList
	if len(blocklist) > 0 {
		for _, a := range s.Configuration().BlockList {
			if strings.HasPrefix(remote, a) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}
	}
	log.T.F("http request: %s from %s", r.URL.String(), helpers.GetRemoteFromReq(r))
	s.Mux.ServeHTTP(w, r)
}

func (s *Server) Shutdown() {
	log.W.Ln("shutting down relay")
	s.Cancel()
	log.W.Ln("closing event store")
	chk.E(s.Store.Close())
	log.W.Ln("shutting down relay listener")
	chk.E(s.HTTPServer.Shutdown(s.Ctx))
}
