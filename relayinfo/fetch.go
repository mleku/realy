package relayinfo

import (
	"net/http"
	"time"

	"realy.lol/context"
	"realy.lol/normalize"
	"realy.lol/units"
	"encoding/json"
	"io"
	"errors"
)

// Fetch fetches the NIP-11 Info.
func Fetch[V st | by](c cx, uv V) (info *T, err er) {
	u := by(uv)
	if _, ok := c.Deadline(); !ok {
		// if no timeout is set, force it to 7 seconds
		var cancel context.F
		c, cancel = context.Timeout(c, 7*time.Second)
		defer cancel()
	}
	u = normalize.HTTPURL(u)
	var req *http.Request
	if req, err = http.NewRequestWithContext(c, http.MethodGet, st(u),
		nil); chk.E(err) {
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
	defer resp.Body.Close()
	b := make(by, units.Mb)
	var n no
	if n, err = resp.Body.Read(b); errors.Is(err, io.EOF) || chk.E(err) {
		return
	}
	b = b[:n]
	info = &T{}
	if err = json.Unmarshal(b, info); chk.E(err) {
		return
	}
	return
}
