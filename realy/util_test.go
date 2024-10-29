package realy

import (
	"net/http"
	"testing"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	eventstore "realy.lol/store"
)

// todo: this needs updating

func startTestRelay(c context.T, t *testing.T, tr *testRelay) *Server {
	t.Helper()
	srv, _ := NewServer(c, tr, "")
	started := make(chan bool)
	go srv.Start("127.0.0.1", 0, "127.0.0.1", 0, started)
	<-started
	return srv
}

type testRelay struct {
	name        S
	storage     eventstore.I
	init        func() E
	onShutdown  func(context.T)
	acceptEvent func(*event.T) bool
}

func (tr *testRelay) Name() S                        { return tr.name }
func (tr *testRelay) Storage(context.T) eventstore.I { return tr.storage }
func (tr *testRelay) Origin() S                      { return "example.com" }
func (tr *testRelay) Init() E {
	if fn := tr.init; fn != nil {
		return fn()
	}
	return nil
}

func (tr *testRelay) OnShutdown(ctx context.T) {
	if fn := tr.onShutdown; fn != nil {
		fn(ctx)
	}
}

func (tr *testRelay) AcceptEvent(c context.T, evt *event.T, hr *http.Request,
	authedPubkey B) bool {
	if fn := tr.acceptEvent; fn != nil {
		return fn(evt)
	}
	return true
}

type testStorage struct {
	init        func() E
	close       func()
	queryEvents func(context.T, *filter.T) ([]*event.T, E)
	deleteEvent func(context.T, *eventid.T) E
	saveEvent   func(context.T, *event.T) E
	countEvents func(context.T, *filter.T) (N, E)
}

func (st *testStorage) Nuke() (err eventstore.E) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Path() eventstore.S {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Init() E {
	if fn := st.init; fn != nil {
		return fn()
	}
	return nil
}

func (st *testStorage) Close() (err E) {
	if fn := st.close; fn != nil {
		fn()
	}
	return
}

func (st *testStorage) QueryEvents(c context.T, f *filter.T) (evs []*event.T, err E) {
	if fn := st.queryEvents; fn != nil {
		return fn(c, f)
	}
	return nil, nil
}

func (st *testStorage) DeleteEvent(c context.T, evt *eventid.T) E {
	if fn := st.deleteEvent; fn != nil {
		return fn(c, evt)
	}
	return nil
}

func (st *testStorage) SaveEvent(c context.T, e *event.T) E {
	if fn := st.saveEvent; fn != nil {
		return fn(c, e)
	}
	return nil
}

func (st *testStorage) CountEvents(ctx context.T, f *filter.T) (N, E) {
	if fn := st.countEvents; fn != nil {
		return fn(ctx, f)
	}
	return 0, nil
}
