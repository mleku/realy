package eventenvelope

import (
	"bufio"
	"bytes"
	"testing"

	"realy.lol/envelopes"
	"realy.lol/event"
	"realy.lol/event/examples"
	"realy.lol/subscription"
)

func TestSubmission(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out by
	var err er
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		rem = rem[:0]
		ea := NewSubmissionWith(ev)
		rem = ea.Marshal(rem)
		c = append(c, rem...)
		var l string
		if l, rem, err = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.Unmarshal(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		out = ea.Marshal(out)
		if !equals(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		c, out = c[:0], out[:0]
	}
}

func TestResult(t *testing.T) {
	scanner := bufio.NewScanner(bytes.NewBuffer(examples.Cache))
	var c, rem, out by
	var err er
	for scanner.Scan() {
		b := scanner.Bytes()
		ev := event.New()
		if _, err = ev.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		var ea *Result
		if ea, err = NewResultWith(subscription.NewStd().String(),
			ev); chk.E(err) {
			t.Fatal(err)
		}
		rem = ea.Marshal(rem)
		c = append(c, rem...)
		var l string
		if l, rem, err = envelopes.Identify(rem); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		if rem, err = ea.Unmarshal(rem); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) != 0 {
			t.Fatalf("some of input remaining after marshal/unmarshal: '%s'",
				rem)
		}
		out = ea.Marshal(out)
		if !equals(out, c) {
			t.Fatalf("mismatched output\n%s\n\n%s\n", c, out)
		}
		rem, c, out = rem[:0], c[:0], out[:0]
	}
}

func TestResult_Unmarshal(t *testing.T) {
	var err er
	evbs := []by{
		// by(`["EVENT",":9",{"id":"8031e5c66be1bb00ee55e4902c74fb37d447b9cddda2e2591d21c9bb30a17f60","kind":10002,"pubkey":"146bda4ec6932830503ee4f8e8b626bd7b3a5784232b8240ba15c8cbff9a07cd","created_at":1733221976,"content":"","tags":[["r","wss://relay.primal.net/"],["r","wss://nostr.mom/"],["r","wss://nos.lol/"],["r","wss://bitcoiner.social/"],["r","wss://relay.nostr.bg/"],["r","wss://nostr.oxtr.dev/"],["r","wss://relay.wellorder.net/"],["r","wss://nostr.wine/"],["r","wss://eden.nostr.land/"],["r","wss://offchain.pub/"],["r","wss://nostr-pub.wellorder.net/"],["r","wss://atlas.nostr.land/"],["r","wss://relay.stoner.com/"],["r","wss://puravida.nostr.land/"],["r","wss://nostr.coinfundit.com/"],["r","wss://relay.plebstr.com/"],["r","wss://relay.snort.social/"],["r","wss://nostr.zebedee.cloud/"],["r","wss://nostr.walletofsatoshi.com/"],["r","wss://purplepag.es/"],["r","wss://relay.nostrich.de/"],["r","wss://brb.io/"],["r","wss://relay.siamstr.com/"],["r","wss://relay.nostr.com.au/"],["r","wss://nostr.milou.lol/"],["r","wss://relay.orangepill.dev/"],["r","wss://nostr.bitcoiner.social/"],["r","wss://relay.nostr.band/"],["r","wss://relay.damus.io/"],["r","wss://relay.0xchat.com/"]],"sig":"7a3a1a16ed30e0f3f64d1ce2425cb68ed6cd383afde16f7422dbe691bc745a85b8c59d3292fb6daad22604699552abad9090b534900a33400ed175a3acdf2521"}]`),
		// by(`["EVENT",":4",{"content":"","created_at":1713250740,"id":"79c22b961e808bb94a4c8b92e681f0c72175878c4f4c3d72c47bb179b0dbfb4e","kind":10002,"pubkey":"43debddd908e677f0b559fff1f5b2d99daf8672eb7f88e5ddf7190c9a65e0ca8","sig":"4e690d173abacf70d0bd56cace795622a7ef0a4b513bc5ebe73c7173fd013003051f430e8d6bf189ce6d46251bb814f55f0487b23da31fb35e82d0610a212c13","tags":[["r","wss://nos.lol/"],["r","wss://nostr.bitcoiner.social/"],["r","wss://relay.nostr.band/"],["r","wss://nostr-pub.wellorder.net/"],["r","wss://offchain.pub/"],["r","wss://relay.mostr.pub/"],["r","wss://purplepag.es/"],["r","wss://relay.damus.io/"],["r","wss://nostr.oxtr.dev/"],["r","wss://xmr.usenostr.org/"],["r","wss://relay.snort.social/"],["r","wss://relay.nostrss.re/"],["r","wss://relay.nostr.directory/"],["r","wss://relayable.org/"],["r","wss://nostr.skitso.business/"],["r","wss://nostr.sidnlabs.nl/"],["r","wss://relay.blackbyte.nl/"],["r","wss://milwaukie.nostr1.com/"],["r","wss://GALAXY13.nostr1.com/"],["r","wss://21ideas.nostr1.com/"],["r","wss://support.nostr1.com/"],["r","wss://pater.nostr1.com/"],["r","wss://ryan.nostr1.com/"],["r","wss://nostr21.com/"]]}]`),
		// by(`["EVENT",":3",{"content":"","created_at":1685471660,"id":"b8cb76ed134c0d245d2ed97ae5fae916d6f727dafa6ca25777caa4d6bdcc6e62","kind":10002,"pubkey":"340e4aef86cd5240aaf5fc550fa3f4291e6a03281a469b6ef34f060f6944033f","sig":"4e1661f8561c820aee150b6e841b309e3772eb7f89123aa7b3a3ffc6e88445fa6c3368d88d002987dd883227908aaaa40eee1819ac53ba632ae4df32a695e7b9","tags":[["r","wss://relay.damus.io"],["r","wss://relay.nostr.band"],["r","wss://btc-italia.online"],["r","wss://bitcoiner.social"],["client","coracle"]]}]`),
		// by(`["EVENT",":2",{"content":"","created_at":1707267379,"id":"fd9b42c626277d763877d526043552e0c01fd7cea34b68a191b6981aeb20cb90","kind":10002,"pubkey":"c5fb6ecc876e0458e3eca9918e370cbcd376901c58460512fe537a46e58c38bb","sig":"a4db470fed2d18a341e7bd4e2b15ab9ed9c2e6c5df9d213d3f9ec1b0d848a18b39358fc31be97e78a2d1e19016ea6ce5603bd1645bc50ccfe76d566a40bf1a2f","tags":[["r","wss://nostr21.com"],["r","wss://blastr.f7z.xyz"],["r","wss://relay.orangepill.dev/"],["r","wss://relay.nostriches.org/"],["r","wss://relayable.org"],["r","wss://relay.nostr.band/"],["r","wss://purplepag.es"],["r","wss://welcome.nostr.wine"],["r","wss://nos.lol"],["r","wss://relay.nostrview.com/"],["r","wss://relay.damus.io"],["r","wss://filter.nostr.wine/npub1chakany8dcz93clv4xgcudcvhnfhdyqutprq2yh72daydevv8zasmuhf02?broadcast=true"],["r","wss://nostr.fmt.wiz.biz"],["r","wss://eden.nostr.land"],["r","wss://nostr.plebchain.org/"],["r","wss://nostr.wine"],["r","wss://atlas.nostr.land"],["r","wss://offchain.pub/"],["r","wss://relay.nostrati.com/"],["r","wss://pyramid.fiatjaf.com"]]}]`),
		by(`["EVENT",":3",{"kind":10002,"id":"6ce2cb371eb16d3ba50dc49c5290d05b84ff5dd9c50b89e3e74ca89af9fc89c4","pubkey":"e58143f793e4bf805a4df6cdc0289e352b3cf08a7b3e6afaaf89dd497bf0f4a6","created_at":1731866484,"tags":[["r","wss://eden.nostr.land"],["r","wss://nos.lol"],["r","wss://nostr.fmt.wiz.biz"],["r","wss://nostr.wine"],["r","wss://relay.damus.io"],["r","wss://relay.mostr.pub"],["r","wss://nostrelites.org"],["r","wss://relay.bitcoinpark.com"],["r","wss://sendit.nosflare.com"],["r","wss://wot.nostr.party"]],"content":"","sig":"eb73a422adaacdf9ded42e49ab09bc1227b06b8d2c286b0775d9e2b532f8cc8e6c32de1eb27d89bc70159469d79473e9a0150b29c618abe8684ad524ee82d941"}]`),
		by(`["EVENT",":2",{"content":"","created_at":1727465947,"id":"47bf33c78d58be44b63bac818e2e19597972937d79c54b84ae0bf2a08622edd2","kind":10002,"pubkey":"34ca937f6e91550633ff4d8381b388b0cca22d212ff8e7b953f0f458cb16e915","sig":"c59260219bf0fbe0767d9157183c83b3e1d89f7689c92892a4f904dce0017f993f84b006cfe0b113d70b0332b55d1d3c9bfc0b78accb6aa9bb4915648a05c572","tags":[["r","wss://nos.lol/"],["r","wss://relay.primal.net/"],["r","wss://relay.nostr.band/"],["r","wss://relay.damus.io/"],["r","wss://relay.nostrplebs.com/"],["r","wss://nostr-pub.wellorder.net/","write"],["r","wss://nostr.nodeofsven.com/"],["r","wss://nostr.einundzwanzig.space/"],["r","wss://nostr.walletofsatoshi.com/"],["r","wss://relay.snort.social/"],["r","wss://nostr.thank.eu/"],["r","wss://nostr.hifish.org/"],["r","wss://pyramid.fiatjaf.com/"],["r","wss://relay.wellorder.net/","read"]]}]`),
	}
	for _, b := range evbs {
		ev := NewResult()
		if _, err = ev.Unmarshal(b); chk.E(err) {
			t.Fatal(err)
		}
		log.I.S(ev.Event)
	}
}
