package realy

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"realy.lol/context"
	"realy.lol/relay"
	ri "realy.lol/relayinfo"
	"realy.lol/store"
)

//go:embed version
var version S

func (s *Server) HandleNIP11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.T.Ln("handling relay information document")
	var info *ri.T
	if informationer, ok := s.relay.(relay.Informationer); ok {
		info = informationer.GetNIP11InformationDocument()
	} else {
		// 1, 11, 42, 70, 86, 9
		supportedNIPs := ri.GetList(
			ri.BasicProtocol,
			ri.EventDeletion,
			ri.RelayInformationDocument,
			ri.GenericTagQueries,
			ri.NostrMarketplace,
			ri.EventTreatment,
			ri.CommandResults,
			ri.ParameterizedReplaceableEvents,
		)
		var auther relay.Authenticator
		if auther, ok = s.relay.(relay.Authenticator); ok && auther.ServiceUrl(r) != "" {
			supportedNIPs = append(supportedNIPs, ri.Authentication.N())
		}
		var storage store.I
		if s.relay.Storage(context.Bg()) != nil {
			if _, ok = storage.(relay.EventCounter); ok {
				supportedNIPs = append(supportedNIPs, ri.CountingResults.N())
			}
		}
		log.T.Ln("supported NIPs", supportedNIPs)
		info = &ri.T{
			Name:        s.relay.Name(),
			Description: "relay powered by the realy framework",
			Nips:        supportedNIPs,
			Software:    "https://realy.lol",
			Version:     version,
			Limitation:  ri.Limits{MaxLimit: s.maxLimit},
		}
	}
	if err := json.NewEncoder(w).Encode(info); chk.E(err) {
	}
}
