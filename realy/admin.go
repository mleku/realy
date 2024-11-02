package realy

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"realy.lol/cmd/realy/app"
	"realy.lol/context"
	"realy.lol/hex"
)

func (s *Server) HandleAdmin(w http.ResponseWriter, r *http.Request) {
	log.I.S(r.Header)
	switch {
	case strings.HasPrefix(r.URL.Path, "/export"):
		log.I.F("export of event data requested on admin port")
		store := s.relay.Storage(context.Bg())
		if strings.Count(r.URL.Path, "/") > 1 {
			split := strings.Split(r.URL.Path, "/")
			// there should be 3 for a valid path, an empty, "export" and the final parameter
			if len(split) != 3 {
				fmt.Fprintf(w, "incorrectly formatted export parameter: '%s'",
					r.URL.Path)
				return
			}
			switch split[2] {
			case "users":
				// todo: naughty reaching through interface here lol... but the relay
				//  implementation does have this feature and another impl may not. Perhaps add
				//  a new interface for grabbing the relay's allowed list, and rename things to
				//  be more clear. And add a method for fetching such a relay's allowed writers.
				if rl, ok := s.relay.(*app.Relay); ok {
					follows := make([]B, 0, len(rl.Followed))
					for f := range rl.Followed {
						follows = append(follows, B(f))
					}
					store.Export(s.Ctx, w, follows...)
				}
			default:
				// this should be a hyphen separated list of hexadecimal pubkey values
				var exportPubkeys []B
				pubkeys := strings.Split(split[2], "-")
				for _, pubkey := range pubkeys {
					// check they are valid hex
					pk, err := hex.Dec(pubkey)
					if err != nil {
						log.E.F("invalid public key '%s' in parameters", pubkey)
						continue
					}
					exportPubkeys = append(exportPubkeys, pk)
				}
				store.Export(s.Ctx, w, exportPubkeys...)
			}
		} else {
			store.Export(s.Ctx, w)
		}
	case strings.HasPrefix(r.URL.Path, "/import"):
		log.I.F("import of event data requested on admin port %s", r.RequestURI)
		store := s.relay.Storage(context.Bg())
		read := io.LimitReader(r.Body, r.ContentLength)
		store.Import(read)
	case strings.HasPrefix(r.URL.Path, "/shutdown"):
		fmt.Fprintf(w, "shutting down")
		defer r.Body.Close()
		s.Shutdown()
	default:
		fmt.Fprintf(w, `todo: make help info page`)
	}
}
