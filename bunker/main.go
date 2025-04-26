package bunker

import (
	"encoding/json"
	"net/url"
	"strings"

	"relay.mleku.dev/chk"
	"relay.mleku.dev/context"
	"relay.mleku.dev/event"
	"relay.mleku.dev/keys"
)

type Request struct {
	ID     string   `json:"id"`
	Method string   `json:"method"`
	Params [][]byte `json:"params"`
}

func (r *Request) String() (s string) {
	var j []byte
	var err error
	if j, err = json.Marshal(r); chk.E(err) {
		return
	}
	return string(j)
}

type Response struct {
	ID     string `json:"id"`
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

func (r *Response) String() (s string) {
	var j []byte
	var err error
	if j, err = json.Marshal(r); chk.E(err) {
		return
	}
	return string(j)
}

type Signer interface {
	GetSession(clientPubkey string) (*Session, bool)
	HandleRequest(context.T, *event.T) (req *Request, resp *Response,
		eventResponse *event.T, err error)
}

type RelayReadWrite struct {
	Read, Write bool
}

func IsValidBunkerURL(input string) bool {
	p, err := url.Parse(input)
	if err != nil {
		return false
	}
	if p.Scheme != "bunker" {
		return false
	}
	if !keys.IsValidPublicKey(p.Host) {
		return false
	}
	if !strings.Contains(p.RawQuery, "relay=") {
		return false
	}
	return true
}
