package realy

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strings"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
	"realy.lol/sha256"
)

func (s *Server) auth(r *http.Request) (authed bo) {
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

func (s *Server) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	fmt.Fprintf(w, "you may have not configured your admin username/password")
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/export"):
		if ok := s.auth(r); !ok {
			s.unauthorized(w)
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
		if ok := s.auth(r); !ok {
			s.unauthorized(w)
			return
		}
		log.I.F("import of event data requested on admin port %s", r.RequestURI)
		sto := s.relay.Storage(context.Bg())
		read := io.LimitReader(r.Body, r.ContentLength)
		sto.Import(read)
	case strings.HasPrefix(r.URL.Path, "/shutdown"):
		if ok := s.auth(r); !ok {
			s.unauthorized(w)
			return
		}
		fprintf(w, "shutting down")
		defer chk.E(r.Body.Close())
		s.Shutdown()
	default:
		fprintf(w, "todo: realy web interface page\n\n")
		s.handleRelayInfo(w, r)
	}
}
