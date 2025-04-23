package realy

import (
	"bytes"
	"errors"
	"fmt"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/errorf"
	"realy.mleku.dev/event"
	"realy.mleku.dev/filter"
	"realy.mleku.dev/kinds"
	"realy.mleku.dev/log"
	"realy.mleku.dev/reason"
	"realy.mleku.dev/store"
	"realy.mleku.dev/tag"
)

func (s *Server) Publish(c context.T, evt *event.T) (err error) {
	sto := s.Storage()
	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
		return nil

	} else if evt.Kind.IsReplaceable() {
		// replaceable event, delete before storing
		var evs []*event.T
		f := filter.New()
		f.Authors = tag.New(evt.Pubkey)
		f.Kinds = kinds.New(evt.Kind)
		evs, err = sto.QueryEvents(c, f)
		if err != nil {
			return fmt.Errorf("failed to query before replacing: %w", err)
		}
		if len(evs) > 0 {
			log.T.F("found %d possible duplicate events", len(evs))
			for _, ev := range evs {
				del := true
				if bytes.Equal(ev.Id, evt.Id) {
					continue
				}
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					log.I.F("not replacing newer replaceable event")
					return errorf.W(string(reason.Invalid.F("not replacing newer replaceable event")))
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
						log.T.C(func() string {
							return fmt.Sprintf("%s\nreplacing\n%s", evt.Serialize(),
								ev.Serialize())
						})
						// replaceable events we don't tombstone when replacing, so if deleted, old
						// versions can be restored
						if err = sto.DeleteEvent(c, ev.EventId(), true); chk.E(err) {
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
		f.Authors = tag.New(evt.Pubkey)
		f.Kinds = kinds.New(evt.Kind)
		log.I.F("filter for parameterized replaceable %v %s", f.Tags.ToStringsSlice(),
			f.Serialize())
		if evs, err = sto.QueryEvents(c, f); err != nil {
			return errorf.E("failed to query before replacing: %w", err)
		}
		if len(evs) > 0 {
			for _, ev := range evs {
				del := true
				err = nil
				log.I.F("maybe replace %s", ev.Serialize())
				if ev.CreatedAt.Int() > evt.CreatedAt.Int() {
					return errorf.D(string(reason.Blocked.F("not replacing newer parameterized replaceable event")))
				}
				// not deleting these events because some clients are retarded and the query
				// will pull the new one but a backup can recover the data of old ones
				if ev.Kind.IsDirectoryEvent() {
					del = false
				}
				evdt := ev.Tags.GetFirst(tag.New("d"))
				evtdt := evt.Tags.GetFirst(tag.New("d"))
				log.I.F("%s != %s %v", evdt.Value(), evtdt.Value(),
					!bytes.Equal(evdt.Value(), evtdt.Value()))
				if !bytes.Equal(evdt.Value(), evtdt.Value()) {
					continue
				}
				if del {
					defer func() {
						if err != nil {
							// something went wrong saving the replacement, so we won't delete
							// the event.
							return
						}
						log.T.C(func() string {
							return fmt.Sprintf("%s\nreplacing\n%s", evt.Serialize(),
								ev.Serialize())
						})
						// replaceable events we don't tombstone when replacing, so if deleted, old
						// versions can be restored
						if err = sto.DeleteEvent(c, ev.EventId(), true); chk.E(err) {
							return
						}
					}()
				}
			}
		}
	}
	if err = sto.SaveEvent(c, evt); chk.E(err) && !errors.Is(err, store.ErrDupEvent) {
		return errorf.E("failed to save: %w", err)
	}
	return
}
