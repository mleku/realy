package realy

import (
	"io"
	"net/http"
	"strings"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
)

func (s *Server) exportHandler(w http.ResponseWriter, r *http.Request) {
	if ok, _ := s.authAdmin(r); !ok {
		s.unauthorized(w, r)
		return
	}
	log.I.F("export of event data requested on admin port")
	sto := s.relay.Storage()
	if strings.Count(r.URL.Path, "/") > 1 {
		split := strings.Split(r.URL.Path, "/")
		if len(split) != 3 {
			fprintf(w, "incorrectly formatted export parameter: '%s'", r.URL.Path)
			return
		}
		switch split[2] {
		case "users":
			if rl, ok := s.relay.(*app.Relay); ok {
				follows := make([][]byte, 0, len(rl.Followed))
				for f := range rl.Followed {
					follows = append(follows, []byte(f))
				}
				sto.Export(s.Ctx, w, follows...)
			}
		default:
			var exportPubkeys [][]byte
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
}

func (s *Server) importHandler(w http.ResponseWriter, r *http.Request) {
	if ok, _ := s.authAdmin(r); !ok {
		s.unauthorized(w, r)
		return
	}
	log.I.F("import of event data requested on admin port %s", r.RequestURI)
	sto := s.relay.Storage()
	read := io.LimitReader(r.Body, r.ContentLength)
	sto.Import(read)
	if realy, ok := s.relay.(*app.Relay); ok {
		realy.ZeroLists()
		realy.CheckOwnerLists(context.Bg())
	}
}

func (s *Server) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	if ok, _ := s.authAdmin(r); !ok {
		s.unauthorized(w, r)
		return
	}
	fprintf(w, "shutting down")
	defer chk.E(r.Body.Close())
	s.Shutdown()
}

func (s *Server) handleNuke(w http.ResponseWriter, r *http.Request) {
	if ok, _ := s.authAdmin(r); !ok {
		s.unauthorized(w, r)
		return
	}
	var err error
	if err = s.relay.Storage().Nuke(); chk.E(err) {
	}
	if realy, ok := s.relay.(*app.Relay); ok {
		realy.ZeroLists()
		realy.CheckOwnerLists(context.Bg())
	}
	fprintf(w, "nuked DB\n")
}

func (s *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fprintf(w, "todo: realy web interface page\n\n")
	s.handleRelayInfo(w, r)
}
