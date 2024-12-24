package app

import (
	"time"
	"realy.lol/kinds"
	"realy.lol/kind"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/tag"
	"strings"
	"realy.lol/ws"
	"realy.lol/relayinfo"
	"net/url"
)

func (r *Relay) Spider() {
	// Don't start the spider if the spider key is not configured, many relays
	// require auth, whether the spider key is allowed, those that do,
	// and are, it must be there and this wraps the toggle together with the
	// configuration neatly.
	if len(r.C.SpiderKey) == 0 {
		return
	}
	// we run at first startup
	r.spider(true)
	// re-run the spider every hour to catch any updates that for whatever
	// reason permitted users may have uploaded to other relays via other
	// clients that may not be sending to us.
	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case <-r.Ctx.Done():
			return
		case <-ticker.C:
			r.spider(false)
		}
	}
}

// RelayKinds are the types of events that we want to search and fetch.
var RelayKinds = &kinds.T{
	K: []*kind.T{
		kind.RelayListMetadata,
		kind.DMRelaysList,
	},
}

// spider is the actual function that does a spider run
func (r *Relay) spider(full bo) {
	log.I.F("spidering")
	if r.SpiderSigner == nil {
		panic("bro the signer still not hear")
	}
	var err er
	var evs event.Ts
	sto := r.Storage()
	// count how many npubs we need to search with
	r.Lock()
	nUsers := len(r.Followed)
	// n := r.MaxLimit / 2
	// we probably want to be conservative with how many we query at once
	// on rando relays, so make `n` small
	n := r.MaxLimit
	nQueries := nUsers / n
	// make the list
	users := make([]st, 0, nUsers)
	for v := range r.Followed {
		users = append(users, v)
	}
	r.Unlock()
	relays := make(map[st]struct{})
	relaysUsed := make(map[st]struct{})
	usersWithRelays := make(map[st]struct{})
	log.I.F("finding relay events")
	f := &filter.T{Kinds: RelayKinds, Authors: tag.New(users...)}
	if evs, err = sto.QueryEvents(r.Ctx, f, true); chk.E(err) {
		// fatal
		return
	}
	for _, ev := range evs {
		relays, usersWithRelays = filterRelays(ev, relays, usersWithRelays)
	}

	log.I.F("making query chunks")
	// break into chunks for each query
	chunks := make([][]st, 0, nQueries)
	// starting from the nearest integer (from the total divided by the number
	// per chunk) we know
	for i := nQueries * n; i > 0; i -= n {
		// take the last segment
		last := users[i:]
		// if it happens to be a round number, don't collect an empty slice
		if len(last) < 1 {
			continue
		}
		chunks = append(chunks, users[i:])
		// snip what we took out of the main slice
		users = users[:i]
	}
	// for i, v := range chunks {
	// 	f := &filter.T{Kinds: RelayKinds, Authors: tag.New(v...)}
	// 	if evs, err = sto.QueryEvents(r.Ctx, f); chk.E(err) {
	// 		// fatal
	// 		return
	// 	}
	// 	log.D.F("%d relay events found %d/%d", len(evs), i, len(chunks))
	// 	for _, ev := range evs {
	// 		relays, usersWithRelays = filterRelays(ev, relays, usersWithRelays)
	// 	}
	// }
	log.I.F("%d relays found in db, of %d users",
		len(relays), len(usersWithRelays))
	log.W.F("****************** starting spider ******************")
	// now spider all these relays for the users, and get even moar relays
	var second bo
	var found no
spide:
	for rely := range relays {
		if found > 2 {
			log.W.F("got events from %d relays queried, "+
				"finishing spider for today", found)
			return
		}
		select {
		case <-r.Ctx.Done():
			var o st
			for v := range relays {
				o += v + "\n"
			}
			// log.I.F("found relays: %d\n%s", len(relays), o)
			log.W.F("shutting down")
			return
		default:
		}
		// fetch the relay info
		var inf *relayinfo.T
		if inf, err = relayinfo.Fetch(r.Ctx, by(rely)); chk.E(err) {
			delete(relays, rely)
			log.I.F("deleted relay %s now %d relays on list",
				rely, len(relays))
			continue spide
		}
		// if !inf.Limitation.AuthRequired {
		// 	continue spide
		// }
		log.I.S(inf)
		var rl *ws.Client
		rl, err = ws.ConnectWithAuth(r.Ctx, rely, r.SpiderSigner)
		if err = rl.Connect(r.Ctx); chk.E(err) {
			// chk.E(rl.Close())
			continue spide
		}
		log.D.F("connected to '%s'", rely)
		relaysUsed[rely] = struct{}{}
		// first get some estimate of how many of these events the relay has, if
		// possible
		var count no
		var average time.Duration
		for i, v := range chunks {
			log.D.F("chunk %d/%d from %s so far: %d relays %d users %d, av response %v",
				i, len(chunks), rely, count, len(relays), len(usersWithRelays),
				average)
			if i > 3 {
				if average > time.Second {
					log.I.F("relay %s is throttling us, move on", rely)
					chk.E(rl.Close())
					continue spide
				}
				found++
			}
			f := &filter.T{
				Kinds:   &kinds.T{K: kind.Directory},
				Authors: tag.New(v...),
			}
			started := time.Now()
			if evs, err = rl.QuerySync(r.Ctx, f); chk.E(err) {
				chk.E(rl.Close())
				continue spide
			}
			average += time.Now().Sub(started)
			average /= 2
			count += len(evs)
			for _, ev := range evs {
				relays, usersWithRelays = filterRelays(ev, relays,
					usersWithRelays)
				if err = r.Storage().SaveEvent(r.Ctx, ev); err != nil {
					continue
				}
			}
		}
		log.I.F("%d found so far in this spider run; "+
			"got %d results from %s", found, count, rely)
		chk.E(rl.Close())
	}
	log.I.F("%d relays found, of %d users",
		len(relays), len(usersWithRelays))
	// filter out the relays we used
	for rely := range relaysUsed {
		delete(relays, rely)
	}
	if !second {
		log.I.F("%d new relays found, spidering these", len(relays))
		second = true
		goto spide
	} else {
		// we only will spider the additional ones found
		o := "relays found:\n"
		for v := range relaysUsed {
			o = v + "\n"
		}
		log.I.F("%s", o)
		return
	}
}

func filterRelays(ev *event.T,
	relays, usersWithRelays map[st]struct{}) (r, u map[st]struct{}) {
	// log.I.S(ev)
	if !(ev.Kind.Equal(kind.RelayListMetadata) ||
		ev.Kind.Equal(kind.DMRelaysList)) {

		return relays, usersWithRelays
	}
	var foundSome bo
	t := ev.Tags.GetAll(tag.New("r"))
next:
	for _, tr := range t.F() {
		v := st(tr.Value())
		if len(v) < 5 {
			continue
		}
		// we only want wss, very often ws:// is not routeable address.
		if !strings.HasPrefix(v, "wss") {
			continue
		}
		// remove ones with extra shit after the relay
		if strings.ContainsAny(v, " \n\r\f\t") {
			for i := range v {
				switch v[i] {
				case ' ', '\n', '\t', '\r', '\f':
					// log.I.F("%s", v)
					// log.I.F("%s", ev.Serialize())
					v = v[:i]
					continue next
				}
			}
		}
		// we don't want URLs with query parameters, mostly nostr.wine
		// these are not interesting. also, if they have @ symbols. or
		// = in case the user didn't put the ? in properly also.
		if strings.Contains(v, "?") ||
			strings.Contains(v, "@") ||
			strings.Contains(v, "=") {
			// log.E.F("%s", v)
			continue
		}

		// this means some kind of parsing error or format error. it shouldn't
		// happen, but this check was here because the client code used to have
		// a concurrency issue with overwriting event bytes.
		if strings.Contains(v, "\"") {
			log.E.F("%s", v)
			continue
		}
		// get rid of the slashes
		if strings.HasSuffix(v, "/") {
			// trim it off
			v = v[:len(v)-1]
		}
		// and lastly, we aren't going to use tor for this, so, nope.
		// also no .local this is not routeable
		if strings.Contains(v, ".onion") ||
			strings.Contains(v, ".local") {
			continue
		}
		// weirdly sometimes there is addresses with multiple mangled
		// protocol things in them, this is a waste of time also.
		if len(strings.Split(v, "//")) > 2 {
			continue
		}
		// some relays have subpaths for specific users, so we will just ignore
		// relay URLs that contain `npub1` and more than 3 `/` characters, which
		// is a match on wss://filter.nostr.wine/npub1xxxxxx
		if strings.Contains(v, "/npub1") &&
			!strings.Contains(v, "//npub1") &&
			strings.Count(v, "/") > 2 {
			continue
		}
		// finally, because people are dumb and don't know that URLs are
		// case-insensitive, standardise them
		v = strings.ToLower(v)
		// penultimate test, does it validate as a URL at all
		_, err := url.Parse(v)
		if chk.E(err) {
			return relays, usersWithRelays
		}
		relays[v] = struct{}{}
		foundSome = true
	}
	if foundSome {
		usersWithRelays[st(ev.PubKey)] = struct{}{}
	}
	return relays, usersWithRelays
}
