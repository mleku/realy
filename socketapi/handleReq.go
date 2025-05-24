package socketapi

import (
	"bytes"
	"errors"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/envelopes/authenvelope"
	"realy.lol/envelopes/closedenvelope"
	"realy.lol/envelopes/eoseenvelope"
	"realy.lol/envelopes/eventenvelope"
	"realy.lol/envelopes/reqenvelope"
	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/log"
	"realy.lol/publish"
	"realy.lol/realy/interfaces"
	"realy.lol/realy/pointers"
	"realy.lol/reason"
	"realy.lol/store"
	"realy.lol/tag"
)

func (a *A) HandleReq(c context.T, req []byte, srv interfaces.Server, aut []byte, remote string) (r []byte) {

	sto := srv.Storage()
	var err error
	var rem []byte
	env := reqenvelope.New()
	if rem, err = env.Unmarshal(req); chk.E(err) {
		return reason.Error.F(err.Error())
	}
	if len(rem) > 0 {
		log.I.F("extra '%s'", rem)
	}
	allowed := env.Filters
	var accepted, modified bool
	authRequired := srv.AuthRequired()
	authRequested := a.Listener.AuthRequested()
	allowed, accepted, modified = srv.AcceptReq(c, a.Listener.Req(), env.Subscription.T,
		env.Filters, []byte(a.Listener.Authed()), remote)
	if !accepted || allowed == nil || modified {
		if authRequired && !authRequested {
			a.Listener.RequestAuth()
			if _, err = a.AuthRequiredResponse(env, remote, aut, reason.AuthRequired); chk.E(err) {
				return
			}
			if !modified {
				return
			}
		}
	}
	var notice []byte
	if allowed != env.Filters {
		defer func() {
			if authRequired && !authRequested {
				a.Listener.RequestAuth()
				if notice, err = a.AuthRequiredResponse(env, remote, aut, reason.AuthRequired); chk.E(err) {
					return
				}
				return
			}
		}()
	}
	if allowed == nil {
		return
	}
	for _, f := range allowed.F {
		var i uint
		if pointers.Present(f.Limit) {
			if *f.Limit == 0 {
				continue
			}
			i = *f.Limit
		}
		if authRequired && f.Kinds.IsPrivileged() {
			if notice, err = a.HandleAuthPrivilege(env, f, a.Listener.AuthedBytes(), remote); chk.E(err) {
				return
			}
		}
		var events event.Ts
		// log.D.F("query from %s %0x,%s", remote, a.Listener.AuthedBytes(), f.Serialize())
		if events, err = sto.QueryEvents(c, f); err != nil {
			log.E.F("eventstore: %v", err)
			if errors.Is(err, badger.ErrDBClosed) {
				return
			}
			continue
		}
		aut := a.Listener.AuthedBytes()
		// remove events from muted authors if we have the authed user's mute list.
		if a.Listener.IsAuthed() {
			a.FilterPrivileged(c, sto, aut, events)
		}
		// remove privileged events as they come through in scrape queries
		if events, notice, err = a.CheckPrivilege(events, f, env, srv, aut, remote); chk.E(err) {
			return
		}
		if len(notice) > 0 {
			return notice
		}
		if len(events) == 0 {
			continue
		}
		if err = a.WriteEvents(events, env, int(i)); chk.E(err) {
		}
		// write out the events to the socket
		for _, ev := range events {
			i--
			if i < 0 {
				break
			}
			var res *eventenvelope.Result
			if res, err = eventenvelope.NewResultWith(env.Subscription.T,
				ev); chk.E(err) {
				return
			}
			if err = res.Write(a.Listener); chk.E(err) {
				return
			}
		}
	}
	if err = eoseenvelope.NewFrom(env.Subscription).Write(a.Listener); chk.E(err) {
		return
	}
	if env.Filters != allowed {
		return
	}
	receiver := make(event.C, 32)
	publish.P.Receive(&W{
		Listener: a.Listener,
		Id:       env.Subscription.String(),
		Receiver: receiver,
		Filters:  env.Filters,
	})
	return
}

func (a *A) HandleAuthPrivilege(env *reqenvelope.T, f *filter.T, aut []byte, remote string) (notice []byte, err error) {
	log.T.F("privileged request\n%s", f.Serialize())
	senders := f.Authors
	receivers := f.Tags.GetAll(tag.New("#p"))
	switch {
	case len(a.Listener.Authed()) == 0:
		if notice, err = a.AuthRequiredResponse(env, remote, aut, reason.AuthRequired); chk.E(err) {
			return
		}
		return
	case senders.Contains(a.Listener.AuthedBytes()) ||
		receivers.ContainsAny([]byte("#p"), tag.New(a.Listener.AuthedBytes())):
		log.T.F("user %0x from %s allowed to query for privileged event",
			a.Listener.AuthedBytes(), remote)
	default:
		notice = reason.Restricted.F("authenticated user %0x does not have authorization for "+
			"requested filters", a.Listener.AuthedBytes())
	}
	return
}

func (a *A) FilterPrivileged(c context.T, sto store.I, aut []byte, events event.Ts) (evs event.Ts) {
	var mutes event.Ts
	var err error
	if mutes, err = sto.QueryEvents(c, &filter.T{Authors: tag.New(aut),
		Kinds: kinds.New(kind.MuteList)}); !chk.E(err) {
		var mutePubs [][]byte
		for _, ev := range mutes {
			for _, t := range ev.Tags.ToSliceOfTags() {
				if bytes.Equal(t.Key(), []byte("p")) {
					var p []byte
					if p, err = hex.Dec(string(t.Value())); chk.E(err) {
						continue
					}
					mutePubs = append(mutePubs, p)
				}
			}
		}
		var tmp event.Ts
		for _, ev := range events {
			for _, pk := range mutePubs {
				if bytes.Equal(ev.Pubkey, pk) {
					continue
				}
				tmp = append(tmp, ev)
			}
		}
		// remove privileged events
		evs = tmp
	}
	return
}

func (a *A) CheckPrivilege(events event.Ts, f *filter.T, env *reqenvelope.T,
	srv interfaces.Server, aut []byte, remote string) (evs event.Ts, notice []byte, err error) {

	authRequired := srv.AuthRequired()
	isPrivileged := f.Kinds.IsPrivileged()
	for _, ev := range events {
		// if auth is required, kind is privileged and there is no authed pubkey, skip
		if authRequired && isPrivileged && len(aut) == 0 {
			log.I.F("privileged and not authed")
			if notice, err = a.AuthRequiredResponse(env, remote, aut, reason.Restricted); chk.E(err) {
				return
			}
			return
		}
		// if the authed pubkey is not present in the pubkey or p tags, skip
		receivers := f.Tags.GetAll(tag.New("#p"))
		if receivers == nil {
			continue
		}
		if isPrivileged && !(bytes.Equal(ev.Pubkey, aut) ||
			!receivers.ContainsAny([]byte("#p"), tag.New(a.Listener.AuthedBytes()))) {

			log.I.F("%v && (%v || %v)", isPrivileged, !bytes.Equal(ev.Pubkey, aut),
				!receivers.ContainsAny([]byte("#p"), tag.New(a.Listener.AuthedBytes())))
			if notice, err = a.AuthRequiredResponse(env, remote, aut, reason.Restricted); chk.E(err) {
				return
			}
			return
		}
		evs = append(evs, ev)
	}
	return
}

func (a *A) WriteEvents(events event.Ts, env *reqenvelope.T, i int) (err error) {
	// write out the events to the socket
	for _, ev := range events {
		i--
		if i < 0 {
			break
		}
		var res *eventenvelope.Result
		if res, err = eventenvelope.NewResultWith(env.Subscription.T,
			ev); chk.E(err) {
			return
		}
		if err = res.Write(a.Listener); chk.E(err) {
			return
		}
	}
	return
}

func (a *A) AuthRequiredResponse(env *reqenvelope.T, remote string, aut []byte, r reason.R) (notice []byte, err error) {
	if err = closedenvelope.NewFrom(env.Subscription,
		r.F(privilegedClosedNotice)).Write(a.Listener); chk.E(err) {
	}
	if len(aut) < 1 {
		log.I.F("requesting auth from client from %s %0x", remote, a.Listener.AuthedBytes())
		if err = authenvelope.NewChallengeWith(a.Listener.Challenge()).Write(a.Listener); chk.E(err) {
			return
		}
	}
	notice = r.F(privilegedNotice)
	return
}

var privilegedNotice = "this realy does not serve DMs or Application Specific Data " +
	"to unauthenticated users or to npubs not found in the event tags or author fields, does your " +
	"client implement NIP-42?"

var privilegedClosedNotice = "auth required for processing request due to presence of privileged kinds (DMs, app specific data)"
