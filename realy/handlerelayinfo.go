package realy

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"realy.lol/relay"
	ri "realy.lol/relayinfo"
	"realy.lol/store"
)

//go:embed version
var version S

func (s *Server) HandleNIP11(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var info *ri.T
	if informationer, ok := s.relay.(relay.Informationer); ok {
		info = informationer.GetNIP11InformationDocument()
	} else {
		supportedNIPs := ri.GetList(
			ri.EventDeletion,
			ri.RelayInformationDocument,
			ri.GenericTagQueries,
			ri.NostrMarketplace,
			ri.EventTreatment,
			ri.CommandResults,
			ri.ParameterizedReplaceableEvents,
		)
		if _, ok = s.relay.(relay.Authenticator); ok {
			supportedNIPs = append(supportedNIPs, ri.Authentication.N())
		}
		var storage store.I
		if storage, ok = s.relay.(store.I); ok && storage != nil {
			if _, ok = storage.(relay.EventCounter); ok {
				supportedNIPs = append(supportedNIPs, ri.CountingResults.N())
			}
		}

		info = &ri.T{
			Name:        s.relay.Name(),
			Description: "relay powered by the realy framework",
			Nips:        supportedNIPs,
			Software:    "https://realy.lol",
			Version:     version,
		}
	}
	if err := json.NewEncoder(w).Encode(info); chk.E(err) {
	}
}
