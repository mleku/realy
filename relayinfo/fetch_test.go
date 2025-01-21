package relayinfo

import (
	"testing"
	"realy.lol/context"
	"time"
)

func TestFetch(t *testing.T) {
	relays := []st{
		"wss://relay.damus.io",
		"wss://relay.nostr.band",
		"wss://nostr.land",
		"wss://nostr.wine",
		"wss://theforest.nostr1.com",
		"wss://mleku.realy.lol",
		"wss://test.realy.lol",
		"wss://nostr.mom",
		"wss://brb.io",
		"wss://nos.lol",
	}
	for _, rely := range relays {
		c, cancel := context.Timeout(context.Bg(), time.Second*3)
		inf, err := Fetch(c, rely)
		if chk.E(err) {
			log.I.F("relay %s dun gib infos", rely)
			cancel()
			continue
		}
		_ = inf
		// log.I.S(inf)
		cancel()
	}
}
