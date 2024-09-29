package store

import (
	"errors"
	"fmt"
	"sort"

	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/kinds"
	"realy.lol/normalize"
	"realy.lol/tag"
	"realy.lol/ws"
)

// RelayInterface is a wrapper thing that unifies Store and Relay under a common
// API.
type RelayInterface interface {
	Publish(c Ctx, evt *event.T) E
	QuerySync(c Ctx, f *filter.T, opts ...ws.SubscriptionOption) ([]*event.T, E)
}

type RelayWrapper struct {
	I
}

var _ RelayInterface = (*RelayWrapper)(nil)

func (w RelayWrapper) Publish(c Ctx, evt *event.T) (err E) {
	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
		return nil

	} else if evt.Kind.IsReplaceable() {
		// replaceable event, delete before storing
		var evs []*event.T
		f := filter.New()
		f.Authors = tag.New(evt.PubKey)
		f.Kinds = kinds.New(evt.Kind)
		evs, err = w.I.QueryEvents(c, f)
		if err != nil {
			return fmt.Errorf("failed to query before replacing: %w", err)
		}
		if len(evs) > 0 {
			for _, ev := range evs {
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					return errorf.W(S(normalize.Invalid.F("not replacing newer event")))
				}
				log.I.F("%s\nreplacing\n%s", evt.Serialize(), ev.Serialize())
				if err = w.I.DeleteEvent(c, ev.EventID()); chk.E(err) {
					continue
				}
			}
		}
	} else if evt.Kind.IsParameterizedReplaceable() {
		log.I.F("parameterized replaceable %s", evt.Serialize())
		// parameterized replaceable event, delete before storing
		var evs []*event.T
		f := filter.New()
		f.Authors = tag.New(evt.PubKey)
		f.Kinds = kinds.New(evt.Kind)
		d := evt.Tags.GetFirst(tag.New("d", ""))
		log.I.S(d)
		log.I.F("filter for parameterized replaceable %s %s", f.Tags, f.Serialize())
		evs, err = w.I.QueryEvents(c, f)
		if err != nil {
			return fmt.Errorf("failed to query before replacing: %w", err)
		}
		if len(evs) > 0 {
			for _, ev := range evs {
				log.I.F("maybe replace %s", ev.Serialize())
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					return errorf.W(S(normalize.Invalid.F("not replacing newer event")))
				}
				evdt := ev.Tags.GetFirst(tag.New("d"))
				evtdt := evt.Tags.GetFirst(tag.New("d"))
				log.I.F("%s != %s", evdt.Value(), evtdt.Value())
				if !equals(evdt.Value(), evtdt.Value()) {
					continue
				}
				log.I.F("%s\nreplacing\n%s", evt.Serialize(), ev.Serialize())
				if err = w.I.DeleteEvent(c, ev.EventID()); chk.E(err) {
					continue
				}
			}
		}
	}
	if err = w.SaveEvent(c, evt); chk.E(err) && !errors.Is(err, ErrDupEvent) {
		return errorf.E("failed to save: %w", err)
	}

	return nil
}

func (w RelayWrapper) QuerySync(c Ctx, f *filter.T,
	opts ...ws.SubscriptionOption) ([]*event.T, E) {

	evs, err := w.I.QueryEvents(c, f)
	if chk.E(err) {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	if f.Limit != nil && *f.Limit > 0 {
		results := make(event.Descending, 0, *f.Limit)
		for _, ev := range evs {
			results = append(results, ev)
		}
		sort.Sort(results)
		return results, nil
	}
	return nil, nil
}
