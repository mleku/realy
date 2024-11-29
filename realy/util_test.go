package realy

import (
	"io"
	"net/http"
	"testing"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/store"
	"realy.lol/units"
)

// todo: this needs updating

func startTestRelay(c context.T, t *testing.T, tr *testRelay) *Server {
	t.Helper()
	srv, _ := NewServer(ServerParams{
		c, func() {}, tr, "", 500 * units.Kb, "", "",
	})
	started := make(chan bool)
	go srv.Start("127.0.0.1", 0, started)
	<-started
	return srv
}

type testRelay struct {
	Ctx
	Cancel      context.F
	name        S
	storage     store.I
	init        func() E
	onShutdown  func(context.T)
	acceptEvent func(*event.T) bool
}

func (tr *testRelay) Name() S                   { return tr.name }
func (tr *testRelay) Storage(context.T) store.I { return tr.storage }
func (tr *testRelay) Origin() S                 { return "example.com" }
func (tr *testRelay) Init() E {
	tr.Ctx, tr.Cancel = context.Cancel(context.Bg())
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

func (tr *testRelay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, origin S,
	authedPubkey B) (ok bool, notice S, after func()) {
	if fn := tr.acceptEvent; fn != nil {
		return fn(evt), "", nil
	}
	return true, "", nil
}

type testStorage struct {
	init        func() E
	close       func()
	queryEvents func(context.T, *filter.T) ([]*event.T, E)
	deleteEvent func(context.T, *eventid.T) E
	saveEvent   func(context.T, *event.T) E
	countEvents func(context.T, *filter.T) (N, bool, E)
}

func (st *testStorage) Import(r io.Reader) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Export(c store.Ctx, w io.Writer, pubkeys ...store.B) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Sync() (err store.E) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Nuke() (err store.E) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Path() store.S {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Init(path S) E {
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

func (st *testStorage) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err E) {
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

func (st *testStorage) CountEvents(ctx context.T, f *filter.T) (N, bool, E) {
	if fn := st.countEvents; fn != nil {
		return fn(ctx, f)
	}
	return 0, false, nil
}
