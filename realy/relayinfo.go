package realy

import (
	"encoding/json"
	"net/http"

	"realy.lol/relay"
	"realy.lol/relayinfo"
	"realy.lol/store"
	"realy.lol/number"
)

func (s *Server) handleRelayInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.T.Ln("handling relay information document")
	var info *relayinfo.T
	if informationer, ok := s.I.(relay.Informationer); ok {
		info = informationer.GetNIP11InformationDocument()
	} else {
		var supportedNIPs number.List
		var auther relay.Authenticator
		if auther, ok = s.I.(relay.Authenticator); ok && auther.ServiceUrl(r) != "" {
			supportedNIPs = append(supportedNIPs, relayinfo.Authentication.N())
		}
		var storage store.I
		if s.I.Storage() != nil {
			if _, ok = storage.(relay.EventCounter); ok {
				supportedNIPs = append(supportedNIPs,
					relayinfo.CountingResults.N())
			}
		}
		supportedNIPs = relayinfo.GetList(
			relayinfo.BasicProtocol,
			relayinfo.EventDeletion,
			relayinfo.RelayInformationDocument,
			relayinfo.GenericTagQueries,
			relayinfo.NostrMarketplace,
			relayinfo.EventTreatment,
			relayinfo.CommandResults,
			relayinfo.ParameterizedReplaceableEvents,
			relayinfo.ProtectedEvents,
		)
		log.T.Ln("supported NIPs", supportedNIPs)
		info = &relayinfo.T{Name: s.I.Name(),
			Description: "nostr relay powered by the realy framework",
			Nips:        supportedNIPs, Software: "https://realy.lol",
			Version: version,
			Limitation: relayinfo.Limits{
				MaxLimit:         &s.maxLimit,
				AuthRequired:     s.authRequired,
				RestrictedWrites: s.authRequired,
			},
			Icon: "https://cdn.satellite.earth/ac9778868fbf23b63c47c769a74e163377e6ea94d3f0f31711931663d035c4f6.png",
		}

	}
	if err := json.NewEncoder(w).Encode(info); chk.E(err) {
	}
}
