package wrapper

import (
	"errors"
	"fmt"
	"sort"

	"realy.lol/event"
	"realy.lol/filter"
	"realy.lol/kinds"
	"realy.lol/normalize"
	"realy.lol/store"
	"realy.lol/tag"
	"realy.lol/ws"
)

// RelayInterface is a wrapper thing that unifies Store and Relay under a common
// API.
type RelayInterface interface {
	Publish(c cx, evt *event.T) er
	QuerySync(c cx, f *filter.T, opts ...ws.SubscriptionOption) ([]*event.T, er)
}

type Relay struct {
	store.I
}

var _ RelayInterface = (*Relay)(nil)

func (w Relay) Publish(c cx, evt *event.T) (err er) {
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
			log.I.F("found events %d", len(evs))
			for _, ev := range evs {
				del := true
				if equals(ev.ID, evt.ID) {
					continue
				}
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					return errorf.W(st(normalize.Invalid.F("not replacing newer replaceable event")))
				}
				// not deleting these events because some clients are retarded and the query
				// will pull the new one but a backup can recover the data of old ones
				if ev.Kind.IsDirectoryEvent() {
					del = false
				}
				// defer the delete until after the save, further down, has completed.
				if del {
					defer func() {
						if err != nil {
							// something went wrong saving the replacement, so we won't delete
							// the event.
							return
						}
						log.T.C(func() st { return fmt.Sprintf("%s\nreplacing\n%s", evt.Serialize(), ev.Serialize()) })
						// replaceable events we don't tombstone when replacing, so if deleted, old
						// versions can be restored
						if err = w.I.DeleteEvent(c, ev.EventID(), true); chk.E(err) {
							return
						}
					}()
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
		log.I.F("filter for parameterized replaceable %v %s", f.Tags.ToStringSlice(),
			f.Serialize())
		if evs, err = w.I.QueryEvents(c, f); err != nil {
			return errorf.E("failed to query before replacing: %w", err)
		}
		if len(evs) > 0 {
			for _, ev := range evs {
				del := true
				err = nil
				log.I.F("maybe replace %s", ev.Serialize())
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					return errorf.D(st(normalize.Blocked.F("not replacing newer parameterized replaceable event")))
				}
				// not deleting these events because some clients are retarded and the query
				// will pull the new one but a backup can recover the data of old ones
				if ev.Kind.IsDirectoryEvent() {
					del = false
				}
				evdt := ev.Tags.GetFirst(tag.New("d"))
				evtdt := evt.Tags.GetFirst(tag.New("d"))
				log.I.F("%s != %s", evdt.Value(), evtdt.Value())
				if !equals(evdt.Value(), evtdt.Value()) {
					continue
				}
				if del {
					defer func() {
						if err != nil {
							// something went wrong saving the replacement, so we won't delete
							// the event.
							return
						}
						log.T.C(func() st { return fmt.Sprintf("%s\nreplacing\n%s", evt.Serialize(), ev.Serialize()) })
						// replaceable events we don't tombstone when replacing, so if deleted, old
						// versions can be restored
						if err = w.I.DeleteEvent(c, ev.EventID(), true); chk.E(err) {
							return
						}
					}()
				}
			}
		}
	}
	if err = w.SaveEvent(c, evt); chk.E(err) && !errors.Is(err, store.ErrDupEvent) {
		return errorf.E("failed to save: %w", err)
	}
	return
}

func (w Relay) QuerySync(c cx, f *filter.T,
	opts ...ws.SubscriptionOption) ([]*event.T, er) {

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
