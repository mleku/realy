package dns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"realy.lol/bech32encoding/pointers"
	"realy.lol/context"
	"realy.lol/keys"
)

var Nip05Regex = regexp.MustCompile(`^(?:([\w.+-]+)@)?([\w_-]+(\.[\w_-]+)+)$`)

type WellKnownResponse struct {
	Names  map[string]string   `json:"names"`
	Relays map[string][]string `json:"relays,omitempty"`
	NIP46  map[string][]string `json:"nip46,omitempty"`
}

func NewWellKnownResponse() *WellKnownResponse {
	return &WellKnownResponse{
		Names:  make(map[string]string),
		Relays: make(map[string][]string),
		NIP46:  make(map[string][]string),
	}
}

func IsValidIdentifier(input string) bool {
	return Nip05Regex.MatchString(input)
}

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

func QueryIdentifier(c context.T, account string) (prf *pointers.Profile, err error) {
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

func Fetch(c context.T, account string) (resp *WellKnownResponse, name string, err error) {
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

func NormalizeIdentifier(account string) string {
	if strings.HasPrefix(account, "_@") {
		return account[2:]
	}
	return account
}

func StringSliceToByteSlice(ss []string) (bs [][]byte) {
	for _, s := range ss {
		bs = append(bs, []byte(s))
	}
	return
}
