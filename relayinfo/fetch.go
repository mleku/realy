package relayinfo

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"realy.lol/context"
	"realy.lol/normalize"
)

// Fetch fetches the NIP-11 Info.
func Fetch(c cx, u by) (info *T, err er) {
	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		c, cancel = context.Timeout(c, 7*time.Second)
		defer cancel()
	}
	u = normalize.URL(u)
	var req *http.Request
	if req, err = http.NewRequestWithContext(c, http.MethodGet, st(u), nil); chk.E(err) {
		return
	}
	// add the NIP-11 header
	req.Header.Add("Accept", "application/nostr+json")
	// send the response
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); chk.E(err) {
		err = errorf.E("request failed: %w", err)
		return
	}
	defer chk.E(resp.Body.Close())
	var b by
	if b, err = io.ReadAll(resp.Body); chk.E(err) {
		return
	}
	info = &T{}
	if err = json.Unmarshal(b, info); chk.E(err) {
		return
	}
	return
}
