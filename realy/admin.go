package realy

import (
	"io"
	"strings"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/realy/api"
)

func (s *Server) exportHandler(h api.H) {
	if ok := s.auth(h.Request); !ok {
		s.unauthorized(h.ResponseWriter)
		return
	}
	log.I.F("export of event data requested on admin port")
	sto := s.relay.Storage()
	if strings.Count(h.URL.Path, "/") > 1 {
		split := strings.Split(h.URL.Path, "/")
		if len(split) != 3 {
			fprintf(h.ResponseWriter, "incorrectly formatted export parameter: '%s'", h.URL.Path)
			return
		}
		switch split[2] {
		case "users":
			if rl, ok := s.relay.(*app.Relay); ok {
				follows := make([][]byte, 0, len(rl.Followed))
				for f := range rl.Followed {
					follows = append(follows, []byte(f))
				}
				sto.Export(s.Ctx, h.ResponseWriter, follows...)
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
			sto.Export(s.Ctx, h.ResponseWriter, exportPubkeys...)
		}
	} else {
		sto.Export(s.Ctx, h.ResponseWriter)
	}
}

func (s *Server) importHandler(h api.H) {
	if ok := s.auth(h.Request); !ok {
		s.unauthorized(h.ResponseWriter)
		return
	}
	log.I.F("import of event data requested on admin port %s", h.RequestURI)
	sto := s.relay.Storage()
	read := io.LimitReader(h.Body, h.ContentLength)
	sto.Import(read)
	if realy, ok := s.relay.(*app.Relay); ok {
		realy.ZeroLists()
		realy.CheckOwnerLists(context.Bg())
	}
}

func (s *Server) shutdownHandler(h api.H) {
	if ok := s.auth(h.Request); !ok {
		s.unauthorized(h.ResponseWriter)
		return
	}
	fprintf(h.ResponseWriter, "shutting down")
	defer chk.E(h.Body.Close())
	s.Shutdown()
}

func (s *Server) defaultHandler(h api.H) {
	fprintf(h.ResponseWriter, "todo: realy web interface page\n\n")
	s.handleRelayInfo(h)
}
