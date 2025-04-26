package bunker

import (
	"context"
	"fmt"

	"relay.mleku.dev/chk"
	"relay.mleku.dev/dns"
	"relay.mleku.dev/errorf"
)

func queryWellKnownNostrJson(ctx context.Context, fullname string) (pubkey string,
	relays []string, err error) {
	var result *dns.WellKnownResponse
	var name string
	if result, name, err = dns.Fetch(ctx, fullname); chk.E(err) {
		return
	}

	var ok bool
	if pubkey, ok = result.Names[name]; !ok {
		return "", nil, fmt.Errorf("no entry found for the '%s' name", name)
	}
	if relays, _ = result.NIP46[pubkey]; !ok {
		err = errorf.E("no bunker relays found for the '%s' name", name)
		return
	}

	return
}
