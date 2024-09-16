package eventstore

import (
	"errors"
	"fmt"
	"sort"

	"mleku.dev/event"
	"mleku.dev/filter"
	"mleku.dev/ws"
)

// RelayInterface is a wrapper thing that unifies Store and Relay under a// common API.
type RelayInterface interface {
	Publish(c Ctx, evt *event.T) E
	QuerySync(c Ctx, f *filter.T,
		opts ...ws.SubscriptionOption) ([]*event.T, E)
}

type RelayWrapper struct {
	I
}

var _ RelayInterface = (*RelayWrapper)(nil)

func (w RelayWrapper) Publish(c Ctx, evt *event.T) (err E) {
	// var ch event.C
	// defer close(ch)
	if evt.Kind.IsEphemeral() {
		// do not store ephemeral events
		return nil
		// todo: rewrite to fit new API
		// } else if evt.Kind.IsReplaceable() {
		// // replaceable event, delete before storing
		// ch, err = w.Store.QueryEvents(c, &filter.T{
		// 	Authors: []string{evt.PubKey},
		// 	Kinds:   kinds.T{evt.Kind},
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to query before replacing: %w", err)
		// }
		// if previous := <-ch; previous != nil && isOlder(previous, evt) {
		// 	if err = w.Store.DeleteEvent(c, previous); err != nil {
		// 		return fmt.Errorf("failed to delete event for replacing: %w", err)
		// 	}
		// }
		// } else if evt.Kind.IsParameterizedReplaceable() {
		// parameterized replaceable event, delete before storing
		// d := evt.Tags.GetFirst([]string{"d", ""})
		// if d != nil {
		// ch, err = w.Store.QueryEvents(c, &filter.T{
		// 	Authors: []string{evt.PubKey},
		// 	Kinds:   kinds.T{evt.Kind},
		// 	Tags:    filter.TagMap{"d": []string{d.Value()}},
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to query before parameterized replacing: %w", err)
		// }
		// if previous := <-ch; previous != nil && isOlder(previous, evt) {
		// 	if err = w.Store.DeleteEvent(c, previous); chk.D(err) {
		// 		return fmt.Errorf("failed to delete event for parameterized replacing: %w", err)
		// 	}
		// }
		// }
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
	n := f.Limit
	if n != 0 {
		results := make(event.Descending, 0, n)
		for _, ev := range evs {
			results = append(results, ev)
		}
		sort.Sort(results)
		return results, nil
	}
	return nil, nil
}
