package realy

import (
	"encoding/json"
	"net/http"
	"sort"

	"realy.lol"
	"realy.lol/chk"
	"realy.lol/log"
	"realy.lol/relayinfo"
)

func (s *Server) HandleRelayInfo(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	log.I.Ln("handling relay information document")
	var info *relayinfo.T
	supportedNIPs := relayinfo.GetList(
		relayinfo.BasicProtocol,
		relayinfo.EncryptedDirectMessage,
		relayinfo.EventDeletion,
		relayinfo.RelayInformationDocument,
		relayinfo.GenericTagQueries,
		relayinfo.NostrMarketplace,
		relayinfo.EventTreatment,
		relayinfo.CommandResults,
		relayinfo.ParameterizedReplaceableEvents,
		relayinfo.ExpirationTimestamp,
		relayinfo.ProtectedEvents,
		relayinfo.RelayListMetadata,
	)
	if s.ServiceURL(r) != "" {
		supportedNIPs = append(supportedNIPs, relayinfo.Authentication.N())
	}
	sort.Sort(supportedNIPs)
	log.T.Ln("supported NIPs", supportedNIPs)
	info = &relayinfo.T{Name: s.Name,
		Description: realy_lol.Description,
		Nips:        supportedNIPs, Software: realy_lol.URL, Version: realy_lol.Version,
		Limitation: relayinfo.Limits{
			MaxLimit:         s.MaxLimit,
			AuthRequired:     s.AuthRequired(),
			RestrictedWrites: !s.PublicReadable() || s.AuthRequired() || len(s.owners) > 0,
		},
		Icon: "https://cdn.satellite.earth/ac9778868fbf23b63c47c769a74e163377e6ea94d3f0f31711931663d035c4f6.png"}
	if err := json.NewEncoder(w).Encode(info); chk.E(err) {
	}
}
