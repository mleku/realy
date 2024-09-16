package relay

import (
	"context"
	"testing"

	"mleku.dev/event"
	"mleku.dev/filter"
	eventstore "mleku.dev/store"
)

func startTestRelay(t *testing.T, tr *testRelay) *Server {
	t.Helper()
	srv, _ := NewServer(tr, "")
	started := make(chan bool)
	go srv.Start("127.0.0.1", 0, started)
	<-started
	return srv
}

type testRelay struct {
	name        S
	storage     eventstore.I
	init        func() E
	onShutdown  func(context.Context)
	acceptEvent func(*event.T) bool
}

func (tr *testRelay) Name() S                              { return tr.name }
func (tr *testRelay) Storage(context.Context) eventstore.I { return tr.storage }

func (tr *testRelay) Init() E {
	if fn := tr.init; fn != nil {
		return fn()
	}
	return nil
}

func (tr *testRelay) OnShutdown(ctx context.Context) {
	if fn := tr.onShutdown; fn != nil {
		fn(ctx)
	}
}

func (tr *testRelay) AcceptEvent(ctx context.Context, e *event.T) bool {
	if fn := tr.acceptEvent; fn != nil {
		return fn(e)
	}
	return true
}

type testStorage struct {
	init        func() E
	close       func()
	queryEvents func(context.Context, *filter.T) (chan *event.T, E)
	deleteEvent func(context.Context, *event.T) E
	saveEvent   func(context.Context, *event.T) E
	countEvents func(context.Context, *filter.T) (int64, E)
}

func (st *testStorage) Init() E {
	if fn := st.init; fn != nil {
		return fn()
	}
	return nil
}

func (st *testStorage) Close() {
	if fn := st.close; fn != nil {
		fn()
	}
}

func (st *testStorage) QueryEvents(ctx context.Context, f *filter.T) (chan *event.T,
	E) {
	if fn := st.queryEvents; fn != nil {
		return fn(ctx, f)
	}
	return nil, nil
}

func (st *testStorage) DeleteEvent(ctx context.Context, evt *event.T) E {
	if fn := st.deleteEvent; fn != nil {
		return fn(ctx, evt)
	}
	return nil
}

func (st *testStorage) SaveEvent(ctx context.Context, e *event.T) E {
	if fn := st.saveEvent; fn != nil {
		return fn(ctx, e)
	}
	return nil
}

func (st *testStorage) CountEvents(ctx context.Context, f *filter.T) (int64, E) {
	if fn := st.countEvents; fn != nil {
		return fn(ctx, f)
	}
	return 0, nil
}
