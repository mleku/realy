package ws

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"realy.lol/context"
	"realy.lol/filter"
	"realy.lol/filters"
	"realy.lol/kind"
	"realy.lol/kinds"
	"realy.lol/simple"
	"realy.lol/tag"
	"realy.lol/tags"
)

const RELAY = "wss://relay.damus.io"

// test if we can fetch a couple of random events
func TestSubscribeBasic(t *testing.T) {
	rl := mustRelayConnect(RELAY)
	defer rl.Close()
	var lim uint = 2
	sub, err := rl.Subscribe(context.Bg(),
		filters.New(&filter.T{Filter: &simple.Filter{Kinds: kinds.New(kind.TextNote)}, Limit: &lim}))
	if err != nil {
		t.Fatalf("subscription failed: %v", err)
		return
	}
	timeout := time.After(5 * time.Second)
	n := 0
	for {
		select {
		case event := <-sub.Events:
			if event == nil {
				t.Fatalf("event is nil: %v", event)
			}
			n++
		case <-sub.EndOfStoredEvents:
			goto end
		case <-rl.Context().Done():
			t.Errorf("connection closed: %v", rl.Context().Err())
			goto end
		case <-timeout:
			t.Errorf("timeout")
			goto end
		}
	}
end:
	if n != 2 {
		t.Fatalf("expected 2 events, got %d", n)
	}
}

// test if we can do multiple nested subscriptions
func TestNestedSubscriptions(t *testing.T) {
	rl := mustRelayConnect(RELAY)
	defer rl.Close()

	n := atomic.Uint32{}
	_ = n
	// fetch 2 replies to a note
	var lim3 uint = 3
	sub, err := rl.Subscribe(context.Bg(),
		filters.New(&filter.T{
			Filter: &simple.Filter{
				Kinds: kinds.New(kind.TextNote),
				Tags: tags.New(tag.New("e",
					"0e34a74f8547e3b95d52a2543719b109fd0312aba144e2ef95cba043f42fe8c5")),
			},
			Limit: &lim3,
		}))
	if err != nil {
		t.Fatalf("subscription 1 failed: %v", err)
		return
	}

	for {
		select {
		case event := <-sub.Events:
			// now fetch author of this
			var lim uint = 1
			sub, err := rl.Subscribe(context.Bg(),
				filters.New(&filter.T{Filter: &simple.Filter{Kinds: kinds.New(kind.ProfileMetadata),
					Authors: tag.New(event.PubKey)}, Limit: &lim}))
			if err != nil {
				t.Fatalf("subscription 2 failed: %v", err)
				return
			}

			for {
				select {
				case <-sub.Events:
					// do another subscription here in "sync" mode, just so we're sure things are not blocking
					rl.QuerySync(context.Bg(), &filter.T{Limit: &lim})

					n.Add(1)
					if n.Load() == 3 {
						// if we get here it means the test passed
						return
					}
				case <-sub.Context.Done():
					goto end
				case <-sub.EndOfStoredEvents:
					sub.Unsub()
				}
			}
		end:
			fmt.Println("")
		case <-sub.EndOfStoredEvents:
			sub.Unsub()
			return
		case <-sub.Context.Done():
			t.Fatalf("connection closed: %v", rl.Context().Err())
			return
		}
	}
}
