package realy

import (
	"bytes"
	"fmt"
	"net/http"

	"realy.lol/httpauth"
)

// func (s *Server) auth(r *http.Request) (authed bool) {
// 	if s.adminUser == "" || s.adminPass == "" {
// 		// disallow this if it hasn't been configured, the default values are empty.
// 		return
// 	}
// 	username, password, ok := r.BasicAuth()
// 	if ok {
// 		usernameHash := sha256.Sum256(by(username))
// 		passwordHash := sha256.Sum256(by(password))
// 		expectedUsernameHash := sha256.Sum256(by(s.adminUser))
// 		expectedPasswordHash := sha256.Sum256(by(s.adminPass))
// 		usernameMatch := subtle.ConstantTimeCompare(usernameHash[:],
// 			expectedUsernameHash[:]) == 1
// 		passwordMatch := subtle.ConstantTimeCompare(passwordHash[:],
// 			expectedPasswordHash[:]) == 1
// 		if usernameMatch && passwordMatch {
// 			return true
// 		}
// 	}
// 	return
// }

func (s *Server) auth(r *http.Request) (authed bool) {
	var valid bool
	var pubkey []byte
	var err error
	if valid, pubkey, err = httpauth.ValidateRequest(r); chk.E(err) {
		return
	}
	if !valid {
		return
	}
	// check admins pubkey list
	for _, v := range s.admins {
		if bytes.Equal(v.Pub(), pubkey) {
			authed = true
			return
		}
	}
	return
}

func (s *Server) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	fmt.Fprintf(w, "your npub is not welcome here\n")
}

func (s *Server) HandleHTTP(h Handler) {
	log.T.S(h.Request.Header)
	Route(h, Paths{
		"/export":   s.exportHandler,
		"/import":   s.importHandler,
		"/shutdown": s.shutdownHandler,
		"":          s.defaultHandler,
	})
}
