package realy

import (
	"io"

	"realy.lol/filter"
)

// handleFilter is a simplified form of the nip-01 REQ filter, as found in
// filter.S.
//
// Because we are using a REST style interface, some parameters, that are short,
// are found in the URL parameters, this is "since", "until" and maybe there can
// be more later like "sort" to reverse for chronological order, or other
// similar things. There will be no limit. Lists of IDs even several tens of
// thousands is not excessively onerous if the client clearly wants them, they
// can use them, and in addition, with the addition of the prefixes.FullIdIndex
// index going from the filter index to finding the full ID is cheap. A client
// can therefore after all, reuse the results of the query at any time later
// since aside from the implicit "until" of "now" the results prior to that are
// still valid, and likely won't be added to since most events are added at the
// same time as their timestamps are generated.
//
// There is no 'ids' field because this is redundant combined with this API as
// anything in that field overrides everything in a filter.
//
// Instead of the JSON containing the "since" and "until" time window
// specifiers, these are in the HTTP path parameters (as they are short and not
// lists)
//
// The limit is also omitted, because this function only returns lists of events
// as an array. Pagination requires query state to be stored and this is
// expensive, it is better to force the client to maintain this. In addition, it
// is relatively cheap iterating indexes to find keys compared to having to
// unmarshal every value that makes an index match. (for this reason this filter
// requires a new index for the whole event ID so the event key does not need to
// be fetched and unmarshalled.
//
// Likewise, there is no 'search' field as this is a definitely distinct API.
// On-demand full text scanning is somewhat practical on a small result set but
// defining what is a "small filter" is not practical.
func (s *Server) handleFilter(h Handler) {
	var err error
	var req []byte
	// decode the request parameters
	if req, err = io.ReadAll(h.Request.Body); chk.E(err) {
		return
	}
	var rem []byte
	f := filter.NewSimple()
	if rem, err = f.Unmarshal(req); chk.E(err) {
		return
	}
	if len(rem) > 0 {
		log.I.S("rem", rem)
	}
	// todo: we also need the URL parameters
}
