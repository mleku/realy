// Package dns is an implementation of the specification of NIP-05, providing
// DNS based verification for nostr identities.
package dns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"realy.mleku.dev/bech32encoding/pointers"
	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/errorf"
	"realy.mleku.dev/keys"
)

// Nip05Regex is an regular expression that matches up with the same pattern as
// an email address.
var Nip05Regex = regexp.MustCompile(`^(?:([\w.+-]+)@)?([\w_-]+(\.[\w_-]+)+)$`)

// WellKnownResponse is the structure of the JSON to be found at
// <url>/.well-known/nostr.json
type WellKnownResponse struct {
	// Names is a list of usernames associated with the DNS identity as in <name>@<domain>
	Names map[string]string `json:"names"`
	// Relays associates one of the public keys from Names to a list of relay URLs
	// that are recommended for that user.
	Relays map[string][]string `json:"relays,omitempty"`
	NIP46  map[string][]string `json:"nip46,omitempty"` // todo: is this obsolete?
}

// NewWellKnownResponse creates a new WellKnownResponse and is required as all
// the fields are maps and need to be allocated.
func NewWellKnownResponse() *WellKnownResponse {
	return &WellKnownResponse{
		Names:  make(map[string]string),
		Relays: make(map[string][]string),
		NIP46:  make(map[string][]string),
	}
}

// IsValidIdentifier verifies that an identifier matches a correct NIP-05
// username@domain
func IsValidIdentifier(input string) bool {
	return Nip05Regex.MatchString(input)
}

// ParseIdentifier searches a string for a valid NIP-05 username@domain
func ParseIdentifier(account string) (name, domain string, err error) {
	res := Nip05Regex.FindStringSubmatch(account)
	if len(res) == 0 {
		return "", "", errorf.E("invalid identifier")
	}
	if res[1] == "" {
		res[1] = "_"
	}
	return res[1], res[2], nil
}

// QueryIdentifier queries a web server from the domain of a NIP-05 DNS
// identifier
func QueryIdentifier(c context.T, account string) (prf *pointers.Profile,
	err error) {

	var result *WellKnownResponse
	var name string
	if result, name, err = Fetch(c, account); chk.E(err) {
		return
	}
	pubkey, ok := result.Names[name]
	if !ok {
		err = errorf.E("no entry for name '%s'", name)
		return
	}
	if !keys.IsValidPublicKey(pubkey) {
		return nil, errorf.E("got an invalid public key '%s'", pubkey)
	}
	var pkb []byte
	if pkb, err = keys.HexPubkeyToBytes(pubkey); chk.E(err) {
		return
	}
	relays, _ := result.Relays[pubkey]
	return &pointers.Profile{
		PublicKey: pkb,
		Relays:    StringSliceToByteSlice(relays),
	}, nil
}

// Fetch parses a DNS identity to find the URL to query for a NIP-05 identity
// verification document.
func Fetch(c context.T, account string) (resp *WellKnownResponse,
	name string, err error) {

	var domain string
	if name, domain, err = ParseIdentifier(account); chk.E(err) {
		err = errorf.E("failed to parse '%s': %w", account, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequestWithContext(c, "GET",
		fmt.Sprintf("https://%s/.well-known/nostr.json?name=%s", domain, name),
		nil); chk.E(err) {

		return resp, name, errorf.E("failed to create a request: %w", err)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request,
			via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var res *http.Response
	if res, err = client.Do(req); chk.E(err) {
		err = errorf.E("request failed: %w", err)
		return
	}
	defer res.Body.Close()
	resp = NewWellKnownResponse()
	b := make([]byte, 65535)
	var n int
	if n, err = res.Body.Read(b); chk.E(err) {
		return
	}
	b = b[:n]
	if err = json.Unmarshal(b, resp); chk.E(err) {
		err = errorf.E("failed to decode json response: %w", err)
	}
	return
}

// NormalizeIdentifier mainly removes the `_@` from the base username so that
// only the domain remains.
func NormalizeIdentifier(account string) string {
	if strings.HasPrefix(account, "_@") {
		return account[2:]
	}
	return account
}

// StringSliceToByteSlice converts a slice of strings to a slice of slices of
// bytes.
func StringSliceToByteSlice(ss []string) (bs [][]byte) {
	for _, s := range ss {
		bs = append(bs, []byte(s))
	}
	return
}
