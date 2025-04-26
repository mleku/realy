package bunker

import (
	"context"
	"fmt"

	"realy.lol/chk"
	"realy.lol/dns"
	"realy.lol/errorf"
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
	if relays, ok = result.NIP46[pubkey]; !ok {
		err = errorf.E("no bunker relays found for the '%s' name", name)
		return
	}

	return
}
